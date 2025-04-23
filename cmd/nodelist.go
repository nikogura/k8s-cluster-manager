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

// nodelistCmd represents the nodelist command
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
		case "aws":
			profile := os.Getenv("AWS_PROFILE")
			dnsManager := cloudflare.NewCloudFlareManager(cfZoneID, cfToken)
			cm, err := aws.NewAWSClusterManager(ctx, clusterName, profile, dnsManager, verbose)
			if err != nil {
				log.Fatalf("Failed creating cluster manager: %s", err)
			}

			// Get the nodes for the cluster
			nodes, err := cm.GetNodes(clusterName)
			if err != nil {
				err = errors.Wrapf(err, "failed getting cluster nodes")
				log.Fatalf("failed listing nodes for cluster %s: %s", clusterName, err)
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

func init() {
	nodeCmd.AddCommand(nodelistCmd)

}
