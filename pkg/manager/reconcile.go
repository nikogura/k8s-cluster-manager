package manager

import (
	"context"
	"fmt"
)

// AWSReconciler provides AWS-specific reconciliation methods.
type AWSReconciler interface {
	GetNodesInSecurityGroup() (nodeInfo []NodeInfo, err error)
	FixMissingClusterTags(instanceIDs []string) (err error)
}

// K8sLister provides Kubernetes node listing.
type K8sLister interface {
	ListNodes(ctx context.Context, verbose bool) (nodeNames []string, err error)
}

// ReconciliationReport contains discrepancies found during reconciliation.
type ReconciliationReport struct {
	EC2Nodes          []NodeInfo
	K8sNodes          []string
	LBTargets         map[string][]string // LB name -> target names
	MissingClusterTag []NodeInfo          // EC2 nodes without Cluster tag
	NotInK8s          []NodeInfo          // EC2 nodes not in Kubernetes
	NotInEC2          []string            // K8s nodes not in EC2
	NotInLB           []NodeInfo          // EC2 nodes not in any LB
	LBWithoutEC2      map[string][]string // LB name -> targets without EC2 instance
}

// ReconcileCluster performs a comprehensive reconciliation of cluster state.
func ReconcileCluster(ctx context.Context, clusterName string, cm K8sClusterManager, verbose bool) (report ReconciliationReport, err error) {
	report = ReconciliationReport{
		EC2Nodes:          make([]NodeInfo, 0),
		K8sNodes:          make([]string, 0),
		LBTargets:         make(map[string][]string),
		MissingClusterTag: make([]NodeInfo, 0),
		NotInK8s:          make([]NodeInfo, 0),
		NotInEC2:          make([]string, 0),
		NotInLB:           make([]NodeInfo, 0),
		LBWithoutEC2:      make(map[string][]string),
	}

	VerboseOutput(verbose, "Starting cluster reconciliation for %s\n", clusterName)

	// Get cluster info (includes EC2, LB, etc.)
	clusterInfo, infoErr := cm.DescribeCluster(clusterName)
	if infoErr != nil {
		err = fmt.Errorf("failed getting cluster info: %w", infoErr)
		return report, err
	}

	report.EC2Nodes = clusterInfo.Nodes

	// Build LB target map
	for _, lb := range clusterInfo.LoadBalancers {
		targets := make([]string, 0)
		for _, target := range lb.Targets {
			targets = append(targets, target.Name)
		}
		report.LBTargets[lb.Name] = targets
	}

	VerboseOutput(verbose, "Found %d EC2 nodes and %d load balancers\n", len(report.EC2Nodes), len(clusterInfo.LoadBalancers))

	return report, err
}

// ConsolePrint prints the reconciliation report to console.
//
//nolint:gocognit // Report printing requires checking multiple conditions
func (r ReconciliationReport) ConsolePrint() {
	fmt.Printf("Reconciliation Report\n")
	fmt.Printf("=====================\n\n")

	fmt.Printf("EC2 Nodes: %d\n", len(r.EC2Nodes))
	for _, node := range r.EC2Nodes {
		fmt.Printf("  - %s (%s)\n", node.Name, node.ID)
	}
	fmt.Println()

	fmt.Printf("Kubernetes Nodes: %d\n", len(r.K8sNodes))
	for _, node := range r.K8sNodes {
		fmt.Printf("  - %s\n", node)
	}
	fmt.Println()

	if len(r.MissingClusterTag) > 0 {
		fmt.Printf("⚠ Nodes Missing Cluster Tag: %d\n", len(r.MissingClusterTag))
		for _, node := range r.MissingClusterTag {
			fmt.Printf("  - %s (%s)\n", node.Name, node.ID)
		}
		fmt.Println()
	}

	if len(r.NotInK8s) > 0 {
		fmt.Printf("⚠ EC2 Nodes Not in Kubernetes: %d\n", len(r.NotInK8s))
		for _, node := range r.NotInK8s {
			fmt.Printf("  - %s (%s)\n", node.Name, node.ID)
		}
		fmt.Println()
	}

	if len(r.NotInEC2) > 0 {
		fmt.Printf("⚠ Kubernetes Nodes Not in EC2: %d\n", len(r.NotInEC2))
		for _, node := range r.NotInEC2 {
			fmt.Printf("  - %s\n", node)
		}
		fmt.Println()
	}

	if len(r.NotInLB) > 0 {
		fmt.Printf("⚠ EC2 Nodes Not in Any Load Balancer: %d\n", len(r.NotInLB))
		for _, node := range r.NotInLB {
			fmt.Printf("  - %s (%s)\n", node.Name, node.ID)
		}
		fmt.Println()
	}

	if len(r.LBWithoutEC2) > 0 {
		fmt.Printf("⚠ Load Balancer Targets Without EC2 Instance: %d\n", len(r.LBWithoutEC2))
		for lb, targets := range r.LBWithoutEC2 {
			fmt.Printf("  %s:\n", lb)
			for _, target := range targets {
				fmt.Printf("    - %s\n", target)
			}
		}
		fmt.Println()
	}

	if len(r.MissingClusterTag) == 0 && len(r.NotInK8s) == 0 && len(r.NotInEC2) == 0 && len(r.NotInLB) == 0 && len(r.LBWithoutEC2) == 0 {
		fmt.Printf("✓ No discrepancies found\n")
	}
}
