package aws

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/talos"
	"github.com/pkg/errors"
)

const (
	maxControlPlaneConcurrent = 1 // Never upgrade more than 1 CP node at a time (etcd quorum)
	defaultMaxConcurrent      = 2 // Default concurrent worker upgrades
	defaultWaitBetween        = 30 * time.Second
	upgradeHealthCheckTimeout = 10 * time.Minute
)

// UpgradeNode upgrades a single node to the specified Talos version.
func (am *AWSClusterManager) UpgradeNode(nodeName string, version string, options manager.UpgradeOptions) (result manager.UpgradeResult, err error) {
	startTime := time.Now()
	result.Version = version

	// Validate version format
	_, valErr := ValidateTalosVersion(version)
	if valErr != nil {
		err = errors.Wrapf(valErr, "invalid version format")
		return result, err
	}

	// Get node info
	nodeInfo, nodeErr := am.GetNode(nodeName)
	if nodeErr != nil {
		err = errors.Wrapf(nodeErr, "failed to get node %s", nodeName)
		return result, err
	}

	// Build cluster node from node info
	node := &AWSNode{
		NodeName:  nodeInfo.Name,
		IPAddress: "", // We'll need to fetch this
		NodeID:    nodeInfo.ID,
	}

	// Fetch IP address
	fetchErr := am.fetchNodeIP(node)
	if fetchErr != nil {
		err = errors.Wrapf(fetchErr, "failed to fetch IP for node %s", nodeName)
		result.NodesFailed = append(result.NodesFailed, manager.UpgradeFailure{
			NodeName: nodeName,
			Error:    err.Error(),
			Phase:    "pre-flight",
		})
		return result, err
	}

	// Get installer image - create a new EC2 client from config for discovery
	ec2Client := ec2.NewFromConfig(am.Config)
	discovery := &AWSImageDiscovery{
		EC2Client: ec2Client,
		Region:    am.Config.Region,
	}

	installerImage := discovery.GetInstallerImage(version)

	manager.VerboseOutput(am.Verbose, "Upgrading node %s to version %s (installer: %s)\n", nodeName, version, installerImage)

	// Dry run check
	if options.DryRun {
		manager.VerboseOutput(am.Verbose, "[DRY RUN] Would upgrade node %s to %s\n", nodeName, installerImage)
		result.NodesUpgraded = append(result.NodesUpgraded, nodeName)
		result.TotalDuration = time.Since(startTime)
		return result, err
	}

	// Execute upgrade
	upgradeErr := talos.UpgradeNode(am.Context, node, installerImage, options.Preserve, options.Stage, am.Verbose)
	if upgradeErr != nil {
		err = errors.Wrapf(upgradeErr, "failed to upgrade node %s", nodeName)
		result.NodesFailed = append(result.NodesFailed, manager.UpgradeFailure{
			NodeName: nodeName,
			Error:    err.Error(),
			Phase:    "upgrade",
		})
		return result, err
	}

	// Wait for node to become healthy (unless staged)
	if !options.Stage {
		waitErr := talos.WaitForNodeHealthy(am.Context, node, upgradeHealthCheckTimeout, am.Verbose)
		if waitErr != nil {
			err = errors.Wrapf(waitErr, "node %s failed health check after upgrade", nodeName)
			result.NodesFailed = append(result.NodesFailed, manager.UpgradeFailure{
				NodeName: nodeName,
				Error:    err.Error(),
				Phase:    "health-check",
			})
			return result, err
		}

		// Verify version
		matches, actualVersion, verifyErr := talos.VerifyNodeVersion(am.Context, node, version, am.Verbose)
		if verifyErr != nil {
			err = errors.Wrapf(verifyErr, "failed to verify version for node %s", nodeName)
			result.NodesFailed = append(result.NodesFailed, manager.UpgradeFailure{
				NodeName: nodeName,
				Error:    err.Error(),
				Phase:    "verification",
			})
			return result, err
		}

		if !matches {
			err = fmt.Errorf("node %s version mismatch: expected %s, got %s", nodeName, version, actualVersion)
			result.NodesFailed = append(result.NodesFailed, manager.UpgradeFailure{
				NodeName: nodeName,
				Error:    err.Error(),
				Phase:    "verification",
			})
			return result, err
		}
	}

	result.NodesUpgraded = append(result.NodesUpgraded, nodeName)
	result.TotalDuration = time.Since(startTime)

	manager.VerboseOutput(am.Verbose, "Successfully upgraded node %s to version %s\n", nodeName, version)

	return result, err
}

// UpgradeCluster orchestrates a rolling upgrade of all cluster nodes.
func (am *AWSClusterManager) UpgradeCluster(version string, options manager.UpgradeOptions) (result manager.UpgradeResult, err error) {
	startTime := time.Now()
	result.Version = version

	manager.VerboseOutput(am.Verbose, "Starting cluster upgrade to version %s\n", version)

	// Validate version
	_, valErr := ValidateTalosVersion(version)
	if valErr != nil {
		err = errors.Wrapf(valErr, "invalid version format")
		return result, err
	}

	// Get all cluster nodes
	clusterInfo, infoErr := am.DescribeCluster(am.ClusterName())
	if infoErr != nil {
		err = errors.Wrapf(infoErr, "failed to describe cluster")
		return result, err
	}

	// Separate control plane and worker nodes
	var cpNodes []string
	var workerNodes []string

	for _, node := range clusterInfo.Nodes {
		if strings.Contains(strings.ToLower(node.Name), "cp") {
			cpNodes = append(cpNodes, node.Name)
		} else {
			workerNodes = append(workerNodes, node.Name)
		}
	}

	// Sort to ensure consistent upgrade order
	sort.Strings(cpNodes)
	sort.Strings(workerNodes)

	manager.VerboseOutput(am.Verbose, "Found %d control plane nodes and %d worker nodes\n", len(cpNodes), len(workerNodes))

	// Determine upgrade order
	var upgradeGroups [][]string

	if options.ControlPlaneFirst || len(cpNodes) > 0 {
		// Always upgrade CP nodes first (one at a time)
		upgradeGroups = append(upgradeGroups, cpNodes)
		upgradeGroups = append(upgradeGroups, workerNodes)
	} else {
		// Upgrade all nodes together
		allNodes := append(cpNodes, workerNodes...)
		upgradeGroups = append(upgradeGroups, allNodes)
	}

	// Execute upgrades
	for groupIdx := range upgradeGroups {
		group := upgradeGroups[groupIdx]
		isControlPlane := groupIdx == 0 && options.ControlPlaneFirst

		maxConcurrent := options.MaxConcurrent
		if maxConcurrent == 0 {
			maxConcurrent = defaultMaxConcurrent
		}

		// Control plane always upgrades one at a time
		if isControlPlane {
			maxConcurrent = maxControlPlaneConcurrent
		}

		groupResult, groupErr := am.upgradeNodeGroup(group, version, options, maxConcurrent)
		if groupErr != nil {
			err = errors.Wrapf(groupErr, "failed upgrading node group")
			// Merge partial results
			result.NodesUpgraded = append(result.NodesUpgraded, groupResult.NodesUpgraded...)
			result.NodesFailed = append(result.NodesFailed, groupResult.NodesFailed...)
			result.TotalDuration = time.Since(startTime)
			return result, err
		}

		// Merge results
		result.NodesUpgraded = append(result.NodesUpgraded, groupResult.NodesUpgraded...)
		result.NodesFailed = append(result.NodesFailed, groupResult.NodesFailed...)

		// Wait between groups
		if groupIdx < len(upgradeGroups)-1 && options.WaitBetween > 0 {
			manager.VerboseOutput(am.Verbose, "Waiting %s before upgrading next group\n", options.WaitBetween)
			time.Sleep(options.WaitBetween)
		}
	}

	result.TotalDuration = time.Since(startTime)

	manager.VerboseOutput(am.Verbose, "Cluster upgrade completed: %d nodes upgraded, %d failed\n",
		len(result.NodesUpgraded), len(result.NodesFailed))

	return result, err
}

//nolint:unparam // err parameter kept for consistency and future extensibility
func (am *AWSClusterManager) upgradeNodeGroup(nodes []string, version string, options manager.UpgradeOptions, maxConcurrent int) (result manager.UpgradeResult, err error) {
	result.Version = version

	// Process nodes in batches
	for i := 0; i < len(nodes); i += maxConcurrent {
		end := i + maxConcurrent
		if end > len(nodes) {
			end = len(nodes)
		}

		batch := nodes[i:end]

		// Upgrade batch
		for _, nodeName := range batch {
			nodeResult, nodeErr := am.UpgradeNode(nodeName, version, options)
			if nodeErr != nil {
				manager.VerboseOutput(am.Verbose, "Failed to upgrade node %s: %v\n", nodeName, nodeErr)
				result.NodesFailed = append(result.NodesFailed, nodeResult.NodesFailed...)
				continue
			}

			result.NodesUpgraded = append(result.NodesUpgraded, nodeResult.NodesUpgraded...)

			// Wait between nodes in batch
			if options.WaitBetween > 0 {
				waitDuration := options.WaitBetween
				if waitDuration == 0 {
					waitDuration = defaultWaitBetween
				}
				manager.VerboseOutput(am.Verbose, "Waiting %s before next node\n", waitDuration)
				time.Sleep(waitDuration)
			}
		}
	}

	return result, err
}

func (am *AWSClusterManager) fetchNodeIP(node *AWSNode) (err error) {
	// Get EC2 instance details
	instances, fetchErr := am.GetEC2InstancesByNodeID(node.NodeID)
	if fetchErr != nil {
		err = errors.Wrapf(fetchErr, "failed to fetch EC2 instance for node %s", node.NodeName)
		return err
	}

	if len(instances) == 0 {
		err = fmt.Errorf("no EC2 instance found for node %s", node.NodeName)
		return err
	}

	if instances[0].PrivateIpAddress == nil {
		err = fmt.Errorf("no private IP address for node %s", node.NodeName)
		return err
	}

	node.IPAddress = *instances[0].PrivateIpAddress

	return err
}
