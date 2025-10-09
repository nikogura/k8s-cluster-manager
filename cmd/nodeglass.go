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

// nodeglassCmd represents the nodeglass command.
//
//nolint:gochecknoglobals // Cobra boilerplate
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

			// Delete Node
			delErr := cm.DeleteNode(nodeName)
			if delErr != nil {
				log.Fatalf("error deleting node %s: %s", nodeName, delErr)
			}

			// TODO Wait for Node Termination

			nodeConfig, ncErr := aws.LoadAWSNodeConfig(nodeBytes)
			if ncErr != nil {
				log.Fatalf("Failed loading node config %s: %s", nodeConfigFile, ncErr)
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
	nodeCmd.AddCommand(nodeglassCmd)
}
