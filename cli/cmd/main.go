package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zero-to-l1/cli/internal/dashboard"
	"github.com/zero-to-l1/cli/internal/launcher"
	"github.com/zero-to-l1/cli/internal/scanner"
)

var rootCmd = &cobra.Command{
	Use:   "zero-to-l1",
	Short: "Zero to Secure L1 - Avalanche L1/Subnet deployment and security scanner platform",
	Long:  `Zero to Secure L1 is an end-to-end platform for deploying Avalanche L1/subnets, running ICM security scans, and monitoring validators.`,
}

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a new Avalanche L1/subnet on Avalanche Fuji Testnet",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &launcher.DefaultRunner{}
		return launcher.ExecuteLaunch(cmd.Context(), runner, "deployments")
	},
}

var fixPluginCmd = &cobra.Command{
	Use:     "fix-plugin [chainName]",
	Aliases: []string{"fix-vm", "build-plugin"},
	Short:   "Build and install RPCChainVM protocol v46 compatible Subnet-EVM plugin for AvalancheGo v1.15.0",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chainName := "zrgchain"
		if len(args) > 0 && args[0] != "" {
			chainName = args[0]
		}
		fixer := &launcher.DefaultPluginFixer{}
		return fixer.FixPlugin(cmd.Context(), chainName, "v1.15.0")
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan <target>",
	Short: "Run Avalanche Teleporter/ICM security vulnerability scanner on Solidity contracts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &scanner.DefaultRunner{}
		return runner.Scan(cmd.Context(), args[0])
	},
}

var dashboardPort int
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the web dashboard backend and frontend monitoring interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &dashboard.DefaultRunner{}
		return runner.Start(cmd.Context(), dashboardPort, 5173)
	},
}

func init() {
	dashboardCmd.Flags().IntVarP(&dashboardPort, "port", "p", 8080, "Port for dashboard backend API")

	rootCmd.AddCommand(launchCmd)
	rootCmd.AddCommand(fixPluginCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(dashboardCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
