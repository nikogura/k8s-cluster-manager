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
	"reflect"
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

		configBytes, patchBytes, nodeBytes, err := ConfigsFromVaultOrFile()
		if err != nil {
			log.Fatalf("Failed getting required node data: %s", err)
		}

		switch cloudProvider {
		case "aws":
			profile := os.Getenv("AWS_PROFILE")
			cm, err := aws.NewAWSClusterManager(ctx, clusterName, profile)
			if err != nil {
				log.Fatalf("Failed creating cluster manager: %s", err)
			}

			nodeConfig, err := aws.LoadAWSNodeConfig(nodeBytes)
			if err != nil {
				log.Fatalf("Failed loading node config %s: %s", nodeConfigFile, err)
			}

			// Error out if we don't get a node config containing data.
			if reflect.DeepEqual(nodeConfig, aws.AWSNodeConfig{}) {
				log.Fatalf("No Node Config.  Cannot continue.")
			}

			// Create Node
			err = cm.CreateNode(nodeName, nodeRole, nodeConfig, configBytes, []string{string(patchBytes)})
			if err != nil {
				log.Fatalf("error creating node %s: %s", nodeName, err)
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

func init() {
	nodeCmd.AddCommand(nodecreateCmd)

}
