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
	"reflect"
)

// nodecreateCmd represents the nodecreate command.
//
//nolint:gochecknoglobals // Cobra boilerplate
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

		configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err := ConfigsFromVaultOrFile()
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

			nodeConfig, ncErr := aws.LoadAWSNodeConfig(nodeBytes)
			if ncErr != nil {
				log.Fatalf("Failed loading node config %s: %s", nodeConfigFile, ncErr)
			}

			// Override Node Type if provided
			if nodeType != "" {
				nodeConfig.InstanceType = nodeType
			}

			// Error out if we don't get a node config containing data.
			if reflect.DeepEqual(nodeConfig, aws.AWSNodeConfig{}) {
				log.Fatalf("No Node Config.  Cannot continue.")
			}

			// Create Node
			createErr := cm.CreateNode(nodeName, nodeRole, nodeConfig, configBytes, []string{string(patchBytes)})
			if createErr != nil {
				log.Fatalf("error creating node %s: %s", nodeName, createErr)
			}

		default:
			log.Fatalf("Cloud provider %q is not yet supported.", cloudProvider)
		}
	},
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	nodeCmd.AddCommand(nodecreateCmd)

}
