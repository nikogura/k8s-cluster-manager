package manager

import (
	"context"
	"fmt"
	"net"
	"time"
)

const NodeRoleCp = "controlplane"
const NodeRoleWorker = "worker"

/*
K8sClusterManager provides cluster management operations.

Needs:
CRUD of EC2 instance
CRUD of talos machine config
CRUD of Cloudflare record
CRUD Target group attachements
Secrets from Vault/env

Commands:

Create - Creates a new node (needs node name and cluster name)  Can we infer name from env?  Should we?  Create node and apply talos config.
Get/Retrieve - return info about node
Delete - Deletes a node (needs name)
Update - Change node size?

Glass(nodeName)  - Terminate EC2 instance, Delete node via kubectl, Create new node of same name, apply talos config.

Need to keep track of how many CRUDs / Glasses we have in flight and limit to 1 at a time.
*/
type K8sClusterManager interface {
	ClusterName() (name string)
	CloudProviderName() (name string)
	K8sProviderName() (name string)
	CreateNode(nodeName string, lbName string) (err error)  // Create a Node, attach it to the LB. Register DNS
	DeleteNode(nodeName string) (err error)                 // Remove Node from LB, Delete it. Remove from DNS
	GetNode(nodeName string) (nodeInfo NodeInfo, err error) // Retrieve Node info
	GetNodes(nodeName string) (nodes []NodeInfo, err error) // Retrieve Node info
	UpdateNode(nodeName string) (err error)                 // Update Node.
	DescribeNode(nodeName string) (info NodeInfo, err error)
	DescribeCluster(clusterName string) (info ClusterInfo, err error)
	DNSManager() (manager DNSManager)
}

type DNSManager interface {
	RegisterNode(ctx context.Context, node ClusterNode, verbose bool) (err error)
	DeregisterNode(ctx context.Context, nodeName string, verbose bool) (err error)
}

type DNSManagerStruct struct{}

func (DNSManagerStruct) RegisterNode(ctx context.Context, node ClusterNode, verbose bool) (err error) {
	return err
}
func (DNSManagerStruct) DeregisterNode(ctx context.Context, nodeName string, verbose bool) (err error) {
	return err
}

// CostEstimator provides cost estimation for compute resources.
type CostEstimator interface {
	// EstimateHourlyCost returns the estimated cost per hour for the given instance type in USD.
	EstimateHourlyCost(instanceType string) (costPerHour float64, err error)
	// EstimateDailyCost returns the estimated cost per day (24 hours) for the given instance type in USD.
	EstimateDailyCost(instanceType string) (costPerDay float64, err error)
}

type ClusterInfo struct {
	Name                       string
	Provider                   string
	Nodes                      []NodeInfo
	LoadBalancers              []LBInfo
	ScheduleWorkloadsOnCPNodes bool     // TODO  How do we keep track of this?
	EstimatedDailyCost         *float64 `json:"estimated_daily_cost,omitempty"` // Optional cost estimate in USD
	TotalVCPUs                 int      `json:"total_vcpus,omitempty"`          // Total vCPUs across all nodes
	TotalMemoryGiB             float64  `json:"total_memory_gib,omitempty"`     // Total memory in GiB across all nodes
}

type NodeInfo struct {
	Name         string
	ID           string
	InstanceType string
	VCPUs        int     `json:"vcpus,omitempty"`      // Number of vCPUs for this instance
	MemoryGiB    float64 `json:"memory_gib,omitempty"` // Memory in GiB for this instance
	DailyCost    float64 `json:"daily_cost,omitempty"` // Estimated daily cost in USD
}

type LBInfo struct {
	Name         string
	IsAPIServer  bool
	Targets      []LBTargetInfo
	TargetGroups []LBTargetGroupInfo
}

type LBTargetGroupInfo struct {
	Name string
	Arn  string
	Port int32
}

type LBTargetInfo struct {
	ID    string
	Name  string
	Port  int32
	State string
}

func (i ClusterInfo) ConsolePrint() {
	// print cluster info
	fmt.Printf("Cluster Info for Cluster %q\nProvider: %s\n", i.Name, i.Provider)

	// iterate over nodes, call ConsolePrint() on each
	fmt.Printf("Nodes: (%d)\n", len(i.Nodes))
	for _, node := range i.Nodes {
		node.ConsolePrint("  ")

	}

	// print resource totals if available
	if i.TotalVCPUs > 0 || i.TotalMemoryGiB > 0 {
		fmt.Printf("Cluster Totals:\n")
		if i.TotalVCPUs > 0 {
			fmt.Printf("  Total vCPUs: %d\n", i.TotalVCPUs)
		}
		if i.TotalMemoryGiB > 0 {
			fmt.Printf("  Total Memory: %.1f GiB\n", i.TotalMemoryGiB)
		}
	}

	// print estimated daily cost if available
	if i.EstimatedDailyCost != nil {
		fmt.Printf("  Estimated Daily Cost: $%.2f\n", *i.EstimatedDailyCost)
	}

	fmt.Printf("Load Balancers: (%d)\n", len(i.LoadBalancers))
	// iterate over load balancers
	for _, lb := range i.LoadBalancers {
		lb.ConsolePrint("  ")

		fmt.Printf("%sTargets:\n", "    ")
		// iterate over targets, call consolePrint on each
		for _, target := range lb.Targets {
			target.ConsolePrint("      ")
		}
	}
}

func (i NodeInfo) ConsolePrint(indent string) {
	if i.InstanceType != "" {
		output := fmt.Sprintf("%s%s (%s)", indent, i.Name, i.InstanceType)

		// Add resource specs if available
		if i.VCPUs > 0 && i.MemoryGiB > 0 {
			output += fmt.Sprintf(" - %d vCPUs, %.1f GiB", i.VCPUs, i.MemoryGiB)
		}

		// Add cost if available
		if i.DailyCost > 0 {
			output += fmt.Sprintf(" - $%.2f/day", i.DailyCost)
		}

		fmt.Printf("%s\n", output)
	} else {
		fmt.Printf("%s%s\n", indent, i.Name)
	}
}

func (i LBInfo) ConsolePrint(indent string) {
	fmt.Printf("%s%s\n", indent, i.Name)
}

func (i LBTargetInfo) ConsolePrint(indent string) {
	fmt.Printf("%s%s:%d State: %s\n", indent, i.Name, i.Port, i.State)
}

type ClusterNode interface {
	Name() (nodeName string) // Name of the node
	Role() (role string)     // Role (cp | worker) of the node
	IP() (ip string)         // IP address of the node
	ID() (id string)         // ID of the node
	Domain() (domain string) // Domain of the node
}

func DialWithRetry(ctx context.Context, network, address string, maxRetries int, delay time.Duration, verbose bool) (conn net.Conn, err error) {
	for i := 0; i <= maxRetries; i++ {
		conn, err = (&net.Dialer{Timeout: delay}).DialContext(ctx, network, address)
		if err == nil {
			return conn, err //nolint:nilerr // err is nil here, which is correct for success
		}
		if i < maxRetries {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return conn, err
			case <-time.After(delay):
				VerboseOutput(verbose, "Connection attempt %d failed: %v.  Retrying.\n", i+1, err)
				continue
			}
		} else {
			err = fmt.Errorf("failed to dial after %d retries: %w", maxRetries+1, err)
			return conn, err
		}
	}
	return conn, err
}

func VerboseOutput(verbose bool, message string, args ...any) {
	if verbose {
		if len(args) == 0 {
			fmt.Printf("%s\n", message)
			return
		}

		msg := fmt.Sprintf(message, args...)
		fmt.Printf("%s\n", msg)
	}
}
