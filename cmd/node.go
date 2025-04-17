/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
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
