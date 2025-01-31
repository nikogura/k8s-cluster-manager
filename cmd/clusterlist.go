/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
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

		switch cloudProvider {
		case "aws":
			profile := os.Getenv("AWS_PROFILE")
			cm, err := aws.NewAWSClusterManager(ctx, profile)
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
