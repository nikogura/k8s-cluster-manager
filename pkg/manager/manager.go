package manager

import "fmt"

/*
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
}

type ClusterInfo struct {
	Name          string
	Provider      string
	Nodes         []NodeInfo
	LoadBalancers []LBInfo
}

type NodeInfo struct {
	Name string
	ID   string
}

type LBInfo struct {
	Name    string
	Targets []LBTargetInfo
}

type LBTargetInfo struct {
	ID    string
	Name  string
	Port  int
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
	fmt.Printf("%s%s\n", indent, i.Name)
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
}
