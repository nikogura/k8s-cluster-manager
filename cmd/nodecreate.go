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
			nodeConfig, err := aws.LoadAWSNodeConfigFromFile(nodeConfigFile)
			if err != nil {
				log.Fatalf("Failed loading config file %s: %s", nodeConfigFile, err)
			}

			// Load the talos config from file or vault
			machineConfigBytes, err := os.ReadFile(machineConfigFile)
			if err != nil {
				log.Fatalf("Failed loading machine config file %s: %s", machineConfigFile, err)
			}

			// Create Node
			err = cm.CreateNode(nodeName, nodeRole, nodeConfig, machineConfigBytes, []string{machineConfigPatchFile})
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

}
