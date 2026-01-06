package main

import (
	"fmt"
	"log"

	"github.com/aserto-dev/topaz/topazd/version"
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
	cmdRun.Flags().StringVarP(
		&flagRunConfigFile,
		"config-file", "c", "",
		"set path of configuration file")
	cmdRun.Flags().StringSliceVarP(
		&flagRunBundleFiles,
		"bundle", "b", []string{},
		"load paths as bundle files or root directories (can be specified more than once)")
	cmdRun.Flags().BoolVarP(
		&flagRunWatchLocalBundles,
		"watch", "w", false,
		"if set, local changes to bundle paths trigger a reload")
	cmdRun.Flags().StringSliceVarP(
		&flagRunIgnorePaths,
		"ignore", "", []string{},
		"set file and directory names to ignore during loading local bundles (e.g., '.*' excludes hidden files)")
	cmdRun.Flags().BoolVarP(
		&flagRunDebug,
		"debug", "", false,
		"start debug service")

	rootCmd.AddCommand(cmdRun)

	_ = cmdRun.MarkFlagRequired("config-file")

	rootCmd.AddCommand(
		versionCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
