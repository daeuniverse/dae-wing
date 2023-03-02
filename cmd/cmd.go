package cmd

import (
	"github.com/spf13/cobra"
)

var (
	Version = "unknown"
	rootCmd = &cobra.Command{
		Use:     "dae-wing [flags] [command [argument ...]]",
		Short:   "dae-wing is a integration solution of dae, API and UI.",
		Long:    `dae-wing is a integration solution of dae, API and UI.`,
		Version: Version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(exportCmd)
}
