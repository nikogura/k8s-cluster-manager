/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// nodedeleteCmd represents the nodedelete command
var nodedeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Kubernetes Node",
	Long: `
Delete a Kubernetes Node
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nodedelete called")
	},
}

func init() {
	nodeCmd.AddCommand(nodedeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodedeleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodedeleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
