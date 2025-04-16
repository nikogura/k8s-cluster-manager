/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// nodecreateCmd represents the nodecreate command
var nodecreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Kubernetes Node",
	Long: `
Create a new Kubernetes Node
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if len(args) > 0 {
			if nodeName == "" {
				nodeName = args[0]
			}
		}

		if len(args) > 1 {
			if nodeRole == "" {
				nodeRole = args[1]
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

			// Load the config
			nodeConfig, err := aws.LoadAWSNodeConfigFromFile(configFile)
			if err != nil {
				log.Fatalf("Failed loading config file %s: %s", configFile, err)
			}

			// Create Node
			err = cm.CreateNode(nodeName, nodeRole, nodeConfig)
			if err != nil {
				log.Fatalf("error deleting node %s: %s", nodeName, err)
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

func init() {
	nodeCmd.AddCommand(nodecreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodecreateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodecreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
