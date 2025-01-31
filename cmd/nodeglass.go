/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		fmt.Println("nodeglass called")
	},
}

func init() {
	nodeCmd.AddCommand(nodeglassCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeglassCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeglassCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
