/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

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
		fmt.Println("nodecreate called")
	},
}

func init() {
	nodeCmd.AddCommand(nodecreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodecreateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodecreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
