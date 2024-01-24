package main

import (
	"fmt"
	"os"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/spf13/cobra"
)

var (
	flagServiceConfigFile string
)

var cmdService = &cobra.Command{
	Use:   "service",
	Short: "service manager",
	Long:  `service manager.`,
}

var cmdServiceInstall = &cobra.Command{
	Use:   "install",
	Short: "install service manager",
	Long:  `install service manager.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.Path(flagServiceConfigFile)
		fmt.Fprintf(os.Stderr, "install service config=%q\n", configPath)

		return nil
	},
}

var cmdServiceRemove = &cobra.Command{
	Use:   "remove",
	Short: "remove service manager",
	Long:  `remove service manager.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.Path(flagServiceConfigFile)
		fmt.Fprintf(os.Stderr, "remove service config=%q\n", configPath)

		return nil
	},
}

// nolint: gochecknoinits
func init() {
	cmdService.AddCommand(cmdServiceInstall)
	cmdService.AddCommand(cmdServiceRemove)
	rootCmd.AddCommand(cmdService)
}
