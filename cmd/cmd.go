package cmd

import (
	"github.com/daeuniverse/dae/common/consts"
	"github.com/spf13/cobra"
)

var (
	Version     = "unknown"
	AppName     = "dae-wing"
	Description = ""
	rootCmd     = &cobra.Command{
		Use:     AppName + " [flags] [command [argument ...]]",
		Short:   Description,
		Long:    Description,
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
	consts.AppName = AppName
	if Description == "" {
		Description = AppName + " is a integration solution of dae, API and UI."
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(exportCmd)
}
