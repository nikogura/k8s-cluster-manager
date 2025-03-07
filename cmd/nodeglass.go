/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// nodeglassCmd represents the nodeglass command
var nodeglassCmd = &cobra.Command{
	Use:   "glass",
	Short: "Deletes and Creates a Kubernetes Node",
	Long: `
Deletes and Creates a Kubernetes Node
Convenience wrapper that calls Delete() and then Create().
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

		switch cloudProvider {
		case "aws":
			profile := os.Getenv("AWS_PROFILE")
			cm, err := aws.NewAWSClusterManager(ctx, clusterName, profile)
			if err != nil {
				log.Fatalf("Failed creating cluster manager: %s", err)
			}

			// Delete Node
			err = cm.DeleteNode(nodeName)
			if err != nil {
				log.Fatalf("error deleting node %s: %s", nodeName, err)
			}

			// Create Node
			err = cm.CreateNode(nodeName)
			if err != nil {
				log.Fatalf("error creating node %s: %s", nodeName, err)
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

func init() {
	nodeCmd.AddCommand(nodeglassCmd)
}
