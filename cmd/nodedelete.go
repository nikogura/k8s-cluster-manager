/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/cloudflare"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// nodedeleteCmd represents the nodedelete command
var nodedeleteCmd = &cobra.Command{
	Use:   "delete <node name>",
	Short: "Delete a Kubernetes Node from a Cluster",
	Long: `
Delete a Kubernetes Node from a Cluster.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if len(args) > 0 {
			if nodeName == "" {
				nodeName = args[0]
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
			role := os.Getenv("AWS_ROLE")
			dnsManager := cloudflare.NewCloudFlareManager(cfZoneID, cfToken)
			cm, err := aws.NewAWSClusterManager(ctx, clusterName, profile, role, dnsManager, verbose)
			if err != nil {
				log.Fatalf("Failed creating cluster manager: %s", err)
			}

			// Delete Node
			err = cm.DeleteNode(nodeName)
			if err != nil {
				log.Fatalf("error deleting node %s: %s", nodeName, err)
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}

	},
}

func init() {
	nodeCmd.AddCommand(nodedeleteCmd)
}
