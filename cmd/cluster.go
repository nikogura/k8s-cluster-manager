/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Operations on Kubernetes Clusters",
	Long: `
Operations on Kubernetes Clusters
`,
	//Run: func(cmd *cobra.Command, args []string) {
	//},
}

func init() {
	rootCmd.AddCommand(clusterCmd)

}
