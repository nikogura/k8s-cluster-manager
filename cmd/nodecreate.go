/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/go-homedir"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"log"
	"os"
	"reflect"
	"strings"

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

		// Talos machine configuration
		var machineConfigBytes []byte

		// if a file is has not been specified, and a secret path has, we'll try to get the data out of vault.
		if machineConfigFile == "" && secretPath != "" {
			machineConfigPath := fmt.Sprintf("%s/cluster-%s-machine-%s", secretPath, clusterName, nodeRole)
			if verbose {
				fmt.Printf("Loading machine config from %s\n", machineConfigPath)
			}

			// Unfortuately, the vault secret gets automatically unmarshalled.
			machineConfigData, err := manager.SecretData(client, machineConfigPath, verbose)
			if err != nil {
				log.Fatalf("Error getting secret at %q: %s", secretPath, err)
			}

			//Marshal it back to YAML, cos that's what the talos sdk expects.
			yamlBytes, err := yaml.Marshal(machineConfigData)
			if err != nil {
				log.Fatalf("Failed secret data to yaml")
			}

			machineConfigBytes = yamlBytes

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
		// This is a little magic as the Talos patch loader expects to get yaml as a string or in a filename, so we only handle the case where there is no patch and we have a secret path.
		if machineConfigPatch == "" && secretPath != "" {
			machineConfigPatchPath := fmt.Sprintf("%s/cluster-%s-patch-%s", secretPath, clusterName, nodeRole)
			if verbose {
				fmt.Printf("Loading machine config patch from %s\n", machineConfigPatchPath)
			}
			machineConfigPatchData, err := manager.SecretData(client, machineConfigPatchPath, verbose)
			if err != nil {
				log.Fatalf("Error getting secret at %q: %s", secretPath, err)
			}

			yamlBytes, err := yaml.Marshal(machineConfigPatchData)
			if err != nil {
				log.Fatalf("Failed converting patch data to yaml")
			}

			machineConfigPatch = fmt.Sprintf("%s", yamlBytes)
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

			var nodeConfig aws.AWSNodeConfig

			// If a config file has not been provided, Load the Node Config from Vault
			if nodeConfigFile == "" && secretPath != "" {
				nodeConfigPath := fmt.Sprintf("%s/cluster-%s-node-%s", secretPath, clusterName, nodeRole)
				if verbose {
					fmt.Printf("Loading node config from %s\n", nodeConfigPath)
				}

				nodeConfigData, err := manager.SecretData(client, nodeConfigPath, verbose)
				if err != nil {
					log.Fatalf("Error getting secret at %q: %s", secretPath, err)
				}

				jsonBytes, err := json.Marshal(nodeConfigData)
				if err != nil {
					log.Fatalf("Failed converting json to  yaml")
				}

				nodeConfig, err = aws.LoadAWSNodeConfig(jsonBytes)
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
	nodeCmd.AddCommand(nodecreateCmd)

}
