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
	"time"
)

//nolint:gochecknoglobals // Cobra boilerplate
var monitorInterval int

// monitorCmd represents the monitor command.
//
//nolint:gochecknoglobals // Cobra boilerplate
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Continuously monitor cluster health",
	Long: `
Continuously monitor cluster health by periodically checking:
- EC2 instances with Cluster tag
- Kubernetes node status
- Load balancer target health
- Discrepancies between systems

The monitor will run indefinitely, checking every interval (default 60 seconds).
Press Ctrl+C to stop monitoring.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if len(args) > 0 {
			if clusterName == "" {
				clusterName = args[0]
			}
		}

		if clusterName == "" {
			log.Fatalf("Cannot monitor without a cluster name")
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

			fmt.Printf("Starting continuous monitoring of cluster %s (interval: %ds)\n", clusterName, monitorInterval)
			fmt.Printf("Press Ctrl+C to stop\n")
			fmt.Println("====================================")
			fmt.Println()

			ticker := time.NewTicker(time.Duration(monitorInterval) * time.Second)
			defer ticker.Stop()

			// Run initial check immediately
			monitorOnce(ctx, cm, clusterName)

			// Then run on interval
			for range ticker.C {
				monitorOnce(ctx, cm, clusterName)
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().IntVarP(&monitorInterval, "interval", "i", 60, "Monitoring interval in seconds")
}

//nolint:gocognit,funlen // Monitoring logic requires multiple checks and reporting
func monitorOnce(ctx context.Context, cm *aws.AWSClusterManager, clusterName string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] Checking cluster health...\n", timestamp)

	// Get cluster info
	clusterInfo, infoErr := cm.DescribeCluster(clusterName)
	if infoErr != nil {
		fmt.Printf("❌ ERROR: Failed getting cluster info: %s\n\n", infoErr)
		return
	}

	// Get K8s nodes
	k8sNodes, k8sErr := kubernetes.ListNodes(ctx, false)
	if k8sErr != nil {
		fmt.Printf("❌ ERROR: Failed listing Kubernetes nodes: %s\n\n", k8sErr)
		return
	}

	// Get nodes potentially missing Cluster tag
	untaggedNodes, untaggedErr := cm.GetNodesInSecurityGroup()
	if untaggedErr != nil {
		fmt.Printf("❌ ERROR: Failed checking for untagged nodes: %s\n\n", untaggedErr)
		return
	}

	// Build maps for comparison
	// Normalize EC2 names by stripping domain suffix for comparison
	ec2Map := make(map[string]bool)
	for _, node := range clusterInfo.Nodes {
		shortName := stripDomainSuffix(node.Name)
		ec2Map[shortName] = true
	}

	k8sMap := make(map[string]bool)
	for _, node := range k8sNodes {
		k8sMap[node] = true
	}

	lbTargetMap := make(map[string]bool)
	unhealthyTargets := make([]string, 0)
	for _, lb := range clusterInfo.LoadBalancers {
		for _, target := range lb.Targets {
			shortName := stripDomainSuffix(target.Name)
			lbTargetMap[shortName] = true
			if target.State != "healthy" {
				unhealthyTargets = append(unhealthyTargets, fmt.Sprintf("%s/%s:%d (%s)", lb.Name, target.Name, target.Port, target.State))
			}
		}
	}

	// Count issues
	issueCount := 0

	// Check for unhealthy targets
	if len(unhealthyTargets) > 0 {
		issueCount++
		fmt.Printf("  ⚠ Unhealthy Load Balancer Targets: %d\n", len(unhealthyTargets))
		for _, target := range unhealthyTargets {
			fmt.Printf("    - %s\n", target)
		}
	}

	// Check for missing Cluster tags
	if len(untaggedNodes) > 0 {
		issueCount++
		fmt.Printf("  ⚠ Instances Missing Cluster Tag: %d\n", len(untaggedNodes))
		for _, node := range untaggedNodes {
			fmt.Printf("    - %s (%s)\n", node.Name, node.ID)
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
		issueCount++
		fmt.Printf("  ⚠ EC2 Instances Not in Kubernetes: %d\n", len(notInK8s))
		for _, name := range notInK8s {
			fmt.Printf("    - %s\n", name)
		}
	}

	// Check for K8s not in EC2
	notInEC2 := make([]string, 0)
	for _, node := range k8sNodes {
		if !ec2Map[node] {
			notInEC2 = append(notInEC2, node)
		}
	}
	if len(notInEC2) > 0 {
		issueCount++
		fmt.Printf("  ⚠ Kubernetes Nodes Not in EC2: %d\n", len(notInEC2))
		for _, name := range notInEC2 {
			fmt.Printf("    - %s\n", name)
		}
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
		issueCount++
		fmt.Printf("  ⚠ EC2 Instances Not in Any Load Balancer: %d\n", len(notInLB))
		for _, name := range notInLB {
			fmt.Printf("    - %s\n", name)
		}
	}

	// Summary
	if issueCount == 0 {
		fmt.Printf("  ✓ All systems healthy - EC2: %d, K8s: %d, LB Targets: %d\n", len(clusterInfo.Nodes), len(k8sNodes), len(lbTargetMap))
	} else {
		fmt.Printf("  Found %d issue(s)\n", issueCount)
	}

	fmt.Println()
}
