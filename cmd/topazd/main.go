package main

import (
	"fmt"
	"log"

	"github.com/aserto-dev/topaz/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "topazd [flags]",
	SilenceErrors: true,
	SilenceUsage:  true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("topazd %s\n", version.GetInfo().Version)
	},
}

func main() {
	rootCmd.AddCommand(
		versionCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
