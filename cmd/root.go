/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals // Cobra boilerplate
var clusterName string

//nolint:gochecknoglobals // Cobra boilerplate
var cloudProvider string

//nolint:gochecknoglobals // Cobra boilerplate
var nodeConfigFile string

//nolint:gochecknoglobals // Cobra boilerplate
var machineConfigFile string

//nolint:gochecknoglobals // Cobra boilerplate
var machineConfigPatch string

//nolint:gochecknoglobals // Cobra boilerplate
var verbose bool

//nolint:gochecknoglobals // Cobra boilerplate
var secretPath string

const cloudProviderAWS = "aws"

// rootCmd represents the base command when called without any subcommands.
//
//nolint:gochecknoglobals // Cobra boilerplate
var rootCmd = &cobra.Command{
	Use:   "k8s-cluster-manager",
	Short: "Manage Nik-style k8s clusters",
	Long: `
Manage Nik-style k8s clusters.

Nik-style clusters do not use LoadBalancer style services generally speaking, but rather configure Layer 4 NLB's to route traffic directly to nodes where they're picked up by Ingress Controllers.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	rootCmd.PersistentFlags().StringVarP(&clusterName, "clustername", "c", "", "Cluster Name")
	rootCmd.PersistentFlags().StringVarP(&cloudProvider, "cloudprovider", "p", cloudProviderAWS, "Cloud provider")
	rootCmd.PersistentFlags().StringVarP(&nodeConfigFile, "nodeconfig", "", "", "Path to node config file")
	rootCmd.PersistentFlags().StringVarP(&machineConfigFile, "machineconfig", "", "", "Path to talos machine config file")
	rootCmd.PersistentFlags().StringVarP(&machineConfigPatch, "machineconfigpatch", "", "", "Path to talos machine config patch file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&secretPath, "secretmount", "m", "", "Vault path for secrets.")
}
