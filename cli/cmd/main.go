package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zero-to-l1",
	Short: "Zero to Secure L1 - Avalanche L1/Subnet deployment and security scanner platform",
	Long:  `Zero to Secure L1 is an end-to-end platform for deploying Avalanche L1/subnets, running ICM security scans, and monitoring validators.`,
}

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a new Avalanche L1/subnet",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

