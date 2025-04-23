/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var nodeName string
var nodeRole string

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Operations on Kubernetes Nodes",
	Long: `
Operations on Kubernetes Nodes
`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	nodeCmd.PersistentFlags().StringVarP(&nodeName, "name", "n", "", "Node Name")
	nodeCmd.PersistentFlags().StringVarP(&nodeRole, "role", "r", "worker", "Node Role")
}

// ConfigsFromVaultOrFile will return byte arrays representing the machine config, patch, and node config, pulled either from Vault (if -m is specified) or
func ConfigsFromVaultOrFile() (configBytes []byte, patchBytes []byte, nodeBytes []byte, cfZoneID string, cfToken string, err error) {
	hd, err := homedir.Dir()
	if err != nil {
		err = errors.Wrapf(err, "unable to look up homedir")
		return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
	}

	var configDataFromSecret manager.ConfigData

	if secretPath != "" {
		tokenFile := fmt.Sprintf("%s/.vault-token", hd)

		tokBytes, err := os.ReadFile(tokenFile)
		if err != nil {
			err = errors.Wrapf(err, "No vault token found at %s", tokenFile)
			return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
		}

		tokString := strings.TrimRight(string(tokBytes), "\n")

		client, err := manager.NewVaultClient(tokString, verbose)
		if err != nil {
			err = errors.Wrapf(err, "failed creating vault client")
			return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
		}

		configDataFromSecret, err = manager.ConfigsFromSecret(client, secretPath, clusterName, nodeRole, cloudProvider, verbose)
		if err != nil {
			err = errors.Wrapf(err, "Failed getting secrets")
			return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
		}
	}

	// if a file is has not been specified, and a secret path has, we'll try to get the data out of vault.
	if machineConfigFile == "" {
		configBytes = configDataFromSecret.TalosMachineConfig
	} else { // Alternately, Load the talos config from a file
		configBytes, err = os.ReadFile(machineConfigFile)
		if err != nil {
			err = errors.Wrapf(err, "Failed loading machine config file %s", machineConfigFile)
			return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
		}
	}

	if len(configBytes) == 0 {
		err = errors.Wrapf(err, "Cannot proceed without a Talos machine configuration.")
		return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
	}

	// Load Talos Machine Config Patch from Vault if a patch has not been provided manually but a secret path has.
	// This is a little magic as the Talos patch loader expects to get yaml as a string or in a filename, so we only handle the case where there is no patch provided, in which case we load it from the secret.
	if machineConfigPatch == "" {
		machineConfigPatch = string(configDataFromSecret.TalosMachineConfigPatch)
	}

	if machineConfigPatch == "" {
		err = errors.Wrapf(err, "Cannot proceed with out a talos machine config patch.")
		return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
	}

	if nodeConfigFile == "" {
		nodeBytes = configDataFromSecret.NodeConfig
	} else {
		configBytes, err = os.ReadFile(nodeConfigFile)
		if err != nil {
			err = errors.Wrapf(err, "Failed loading node config file %s", machineConfigFile)
			return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
		}
	}

	if len(nodeBytes) == 0 {
		err = errors.Wrapf(err, "Cannot proceed without a nodeconfiguration.")
		return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err
	}

	cfZoneID = os.Getenv(manager.CLOUDFLARE_ZONE_ID_ENV_VAR)
	if cfZoneID == "" {
		cfZoneID = configDataFromSecret.CloudflareZoneID
	}

	cfToken = os.Getenv(manager.CLOUDFLARE_API_TOKEN_ENV_VAR)
	if cfToken == "" {
		cfToken = configDataFromSecret.CloudflareApiToken
	}

	return configBytes, patchBytes, nodeBytes, cfZoneID, cfToken, err

}
