/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
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

// clusterlistCmd represents the clusterlist command
var clusterlistCmd = &cobra.Command{
	Use:   "list [cluster-name]",
	Short: "List information about a K8S Cluster",
	Long: `
List information about a K8S Cluster
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
			role := os.Getenv("AWS_ROLE")
			dnsManager := cloudflare.NewCloudFlareManager(cfZoneID, cfToken)
			cm, err := aws.NewAWSClusterManager(ctx, clusterName, profile, role, dnsManager, verbose)
			if err != nil {
				log.Fatalf("Failed creating cluster manager: %s", err)
			}

			info, err := cm.DescribeCluster(clusterName)
			if err != nil {
				log.Fatalf("Failed describing cluster: %s", err)
			}

			info.ConsolePrint()

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterlistCmd)

}
