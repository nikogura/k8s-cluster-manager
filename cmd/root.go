/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var clusterName string
var cloudProvider string
var nodeConfigFile string
var machineConfigFile string
var machineConfigPatchFile string

// rootCmd represents the base command when called without any subcommands
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

func init() {
	rootCmd.PersistentFlags().StringVarP(&clusterName, "clustername", "c", "", "Cluster Name")
	rootCmd.PersistentFlags().StringVarP(&cloudProvider, "cloudprovider", "p", "aws", "Cloud provider")
	rootCmd.PersistentFlags().StringVarP(&nodeConfigFile, "nodeconfig", "", "", "Path to node config file")
	rootCmd.PersistentFlags().StringVarP(&machineConfigFile, "machineconfig", "", "", "Path to talos machine config file")
	rootCmd.PersistentFlags().StringVarP(&machineConfigPatchFile, "machineconfigpatch", "", "", "Path to talos machine config patch file")
}
