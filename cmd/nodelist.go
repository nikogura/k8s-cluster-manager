/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/cloudflare"
	"github.com/pkg/errors"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// nodelistCmd represents the nodelist command.
//
//nolint:gochecknoglobals // Cobra boilerplate
var nodelistCmd = &cobra.Command{
	Use:   "list <cluster name>",
	Short: "List Nodes in a cluster",
	Long: `
List nodes in a cluster.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if len(args) > 0 {
			if clusterName == "" {
				clusterName = args[0]
			}
		}

		if clusterName == "" {
			log.Fatalf("Cannot list without a cluster name")
		}

		_, _, _, cfZoneID, cfToken, err := ConfigsFromVaultOrFile()
		if err != nil {
			log.Fatalf("Failed getting required node data: %s", err)
		}

		switch cloudProvider {
		case cloudProviderAWS:
			profile := os.Getenv("AWS_PROFILE")
			role := os.Getenv("AWS_ROLE")
			dnsManager := cloudflare.NewCloudFlareManager(cfZoneID, cfToken)
			cm, cmErr := aws.NewAWSClusterManager(ctx, clusterName, profile, role, dnsManager, verbose)
			if cmErr != nil {
				log.Fatalf("Failed creating cluster manager: %s", cmErr)
			}

			// Get the nodes for the cluster
			nodes, nodesErr := cm.GetNodes(clusterName)
			if nodesErr != nil {
				nodesErr = errors.Wrapf(nodesErr, "failed getting cluster nodes")
				log.Fatalf("failed listing nodes for cluster %s: %s", clusterName, nodesErr)
			}

			fmt.Printf("Nodes: (%d)\n", len(nodes))
			for _, node := range nodes {
				node.ConsolePrint("")
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	nodeCmd.AddCommand(nodelistCmd)

}
