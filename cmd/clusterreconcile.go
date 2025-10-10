/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/cloudflare"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/kubernetes"
	"github.com/spf13/cobra"
	"log"
	"os"
)

//nolint:gochecknoglobals // Cobra boilerplate
var fixTags bool

// clusterreconcileCmd represents the clusterreconcile command.
//
//nolint:gochecknoglobals // Cobra boilerplate
var clusterreconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Reconcile cluster state and fix discrepancies",
	Long: `
Reconcile cluster state by comparing EC2 instances, Kubernetes nodes, and load balancer targets.

This command will:
- List all EC2 instances (with and without Cluster tag)
- List all Kubernetes nodes
- List all load balancer targets
- Report any discrepancies
- Optionally fix missing Cluster tags with --fix-tags
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if len(args) > 0 {
			if clusterName == "" {
				clusterName = args[0]
			}
		}

		if clusterName == "" {
			log.Fatalf("Cannot reconcile without a cluster name")
		}

		switch cloudProvider {
		case cloudProviderAWS:
			profile := os.Getenv("AWS_PROFILE")
			role := os.Getenv("AWS_ROLE")
			cfZoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
			cfAPIToken := os.Getenv("CLOUDFLARE_API_TOKEN")

			dnsManager := cloudflare.NewCloudFlareManager(cfZoneID, cfAPIToken)
			cm, cmErr := aws.NewAWSClusterManager(ctx, clusterName, profile, role, dnsManager, verbose)
			if cmErr != nil {
				log.Fatalf("Failed creating cluster manager: %s", cmErr)
			}

			// Get cluster info
			fmt.Printf("Reconciling cluster %s\n", clusterName)
			fmt.Println("====================================")
			fmt.Println()

			clusterInfo, infoErr := cm.DescribeCluster(clusterName)
			if infoErr != nil {
				log.Fatalf("Failed getting cluster info: %s", infoErr)
			}

			// Get K8s nodes
			k8sNodes, k8sErr := kubernetes.ListNodes(ctx, verbose)
			if k8sErr != nil {
				log.Fatalf("Failed listing Kubernetes nodes: %s", k8sErr)
			}

			// Get nodes potentially missing Cluster tag
			untaggedNodes, untaggedErr := cm.GetNodesInSecurityGroup()
			if untaggedErr != nil {
				log.Fatalf("Failed checking for untagged nodes: %s", untaggedErr)
			}

			// Build maps for comparison
			// Normalize EC2 names by stripping domain suffix for comparison
			ec2Map := make(map[string]bool)
			ec2FullNameMap := make(map[string]string) // short name -> full name
			for _, node := range clusterInfo.Nodes {
				shortName := stripDomainSuffix(node.Name)
				ec2Map[shortName] = true
				ec2FullNameMap[shortName] = node.Name
			}

			k8sMap := make(map[string]bool)
			for _, node := range k8sNodes {
				k8sMap[node] = true
			}

			lbTargetMap := make(map[string]bool)
			for _, lb := range clusterInfo.LoadBalancers {
				for _, target := range lb.Targets {
					shortName := stripDomainSuffix(target.Name)
					lbTargetMap[shortName] = true
				}
			}

			// Report findings
			fmt.Printf("EC2 Instances (with Cluster=%s tag): %d\n", clusterName, len(clusterInfo.Nodes))
			fmt.Printf("Kubernetes Nodes: %d\n", len(k8sNodes))
			fmt.Printf("Load Balancer Targets: %d\n", len(lbTargetMap))
			fmt.Println()

			// Check for missing Cluster tags
			if len(untaggedNodes) > 0 {
				fmt.Printf("⚠ Instances Missing Cluster Tag: %d\n", len(untaggedNodes))
				instanceIDs := make([]string, 0)
				for _, node := range untaggedNodes {
					fmt.Printf("  - %s (%s) %s\n", node.Name, node.ID, node.InstanceType)
					instanceIDs = append(instanceIDs, node.ID)
				}
				fmt.Println()

				if fixTags {
					fmt.Println("Fixing missing Cluster tags...")
					tagErr := cm.FixMissingClusterTags(instanceIDs)
					if tagErr != nil {
						log.Fatalf("Failed fixing tags: %s", tagErr)
					}
					fmt.Printf("✓ Added Cluster=%s tag to %d instances\n", clusterName, len(instanceIDs))
					fmt.Println()
				} else {
					fmt.Println("Run with --fix-tags to automatically add Cluster tag to these instances")
					fmt.Println()
				}
			}

			// Check for EC2 not in K8s
			notInK8s := make([]string, 0)
			for _, node := range clusterInfo.Nodes {
				shortName := stripDomainSuffix(node.Name)
				if !k8sMap[shortName] {
					notInK8s = append(notInK8s, node.Name)
				}
			}
			if len(notInK8s) > 0 {
				fmt.Printf("⚠ EC2 Instances Not in Kubernetes: %d\n", len(notInK8s))
				for _, name := range notInK8s {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println()
			}

			// Check for K8s not in EC2
			notInEC2 := make([]string, 0)
			for _, node := range k8sNodes {
				if !ec2Map[node] {
					notInEC2 = append(notInEC2, node)
				}
			}
			if len(notInEC2) > 0 {
				fmt.Printf("⚠ Kubernetes Nodes Not in EC2: %d\n", len(notInEC2))
				for _, name := range notInEC2 {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println()
			}

			// Check for EC2 not in any LB
			notInLB := make([]string, 0)
			for _, node := range clusterInfo.Nodes {
				shortName := stripDomainSuffix(node.Name)
				if !lbTargetMap[shortName] {
					notInLB = append(notInLB, node.Name)
				}
			}
			if len(notInLB) > 0 {
				fmt.Printf("⚠ EC2 Instances Not in Any Load Balancer: %d\n", len(notInLB))
				for _, name := range notInLB {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println()
			}

			// Check for LB targets not in EC2
			lbWithoutEC2 := make([]string, 0)
			for target := range lbTargetMap {
				if !ec2Map[target] {
					lbWithoutEC2 = append(lbWithoutEC2, target)
				}
			}
			if len(lbWithoutEC2) > 0 {
				fmt.Printf("⚠ Load Balancer Targets Not in EC2: %d\n", len(lbWithoutEC2))
				for _, name := range lbWithoutEC2 {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println()
			}

			// Summary
			if len(untaggedNodes) == 0 && len(notInK8s) == 0 && len(notInEC2) == 0 && len(notInLB) == 0 && len(lbWithoutEC2) == 0 {
				fmt.Println("✓ No discrepancies found - cluster is in sync")
			} else {
				fmt.Println("Reconciliation complete - see warnings above for discrepancies")
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	clusterCmd.AddCommand(clusterreconcileCmd)
	clusterreconcileCmd.Flags().BoolVar(&fixTags, "fix-tags", false, "Automatically fix missing Cluster tags")
}
