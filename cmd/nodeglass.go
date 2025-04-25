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

		configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err := ConfigsFromVaultOrFile()
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

			// TODO Wait for Node Termination

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
	nodeCmd.AddCommand(nodeglassCmd)
}
