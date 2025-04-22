/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"github.com/spf13/cobra"
	"log"
	"os"
	"reflect"
	"strings"
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

		hd, err := homedir.Dir()
		if err != nil {
			log.Fatalf("unable to look up homedir: %s", err)
		}

		tokenFile := fmt.Sprintf("%s/.vault-token", hd)

		tokBytes, err := os.ReadFile(tokenFile)
		if err != nil {
			log.Fatalf("No vault token found at %s: %s", tokenFile, err)
		}

		tokString := strings.TrimRight(string(tokBytes), "\n")

		client, err := manager.NewVaultClient(tokString, verbose)
		if err != nil {
			log.Fatalf("failed creating vault client: %s.", err)
		}

		var configBytesFromSecret, patchBytesFromSecret, nodeBytesFromSecret []byte

		// Talos machine configuration
		var machineConfigBytes []byte

		if secretPath != "" {
			configBytesFromSecret, patchBytesFromSecret, nodeBytesFromSecret, err = manager.ConfigsFromSecret(client, secretPath, clusterName, nodeRole, cloudProvider, verbose)
			if err != nil {
				log.Fatalf("Failed getting secrets: %s", err)
			}
		}

		// if a file is has not been specified, and a secret path has, we'll try to get the data out of vault.
		if machineConfigFile == "" {
			machineConfigBytes = configBytesFromSecret
		} else { // Alternately, Load the talos config from a file
			machineConfigBytes, err = os.ReadFile(machineConfigFile)
			if err != nil {
				log.Fatalf("Failed loading machine config file %s: %s", machineConfigFile, err)
			}
		}

		if len(machineConfigBytes) == 0 {
			log.Fatalf("Cannot proceed without a Talos machine configuration.")
		}

		// Load Talos Machine Config Patch from Vault if a patch has not been provided manually but a secret path has.
		// This is a little magic as the Talos patch loader expects to get yaml as a string or in a filename, so we only handle the case where there is no patch provided, in which case we load it from the secret.
		if machineConfigPatch == "" {
			machineConfigPatch = string(patchBytesFromSecret)
		}

		if machineConfigPatch == "" {
			log.Fatalf("Cannot proceed with out a talos machine config patch.")
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

			// TODO Wait for Node Termination

			var nodeConfig aws.AWSNodeConfig

			// If a config file has not been provided, Load the Node Config from Vault
			if nodeConfigFile == "" {
				nodeConfig, err = aws.LoadAWSNodeConfig(nodeBytesFromSecret)
				if err != nil {
					log.Fatalf("Failed loading node config: %s", err)
				}

			} else { // Otherwise load it from a file
				nodeConfig, err = aws.LoadAWSNodeConfigFromFile(nodeConfigFile)
				if err != nil {
					log.Fatalf("Failed loading config file %s: %s", nodeConfigFile, err)
				}
			}

			// Error out if we don't get a node config containing data.
			if reflect.DeepEqual(nodeConfig, aws.AWSNodeConfig{}) {
				log.Fatalf("No Node Config.  Cannot continue.")
			}

			// Create Node
			err = cm.CreateNode(nodeName, nodeRole, nodeConfig, machineConfigBytes, []string{machineConfigPatch})
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
