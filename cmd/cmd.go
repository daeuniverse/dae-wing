/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package cmd

import (
	"github.com/daeuniverse/dae-wing/db"
	daeConsts "github.com/daeuniverse/dae/common/consts"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     db.AppName + " [flags] [command [argument ...]]",
		Short:   db.AppDescription,
		Long:    db.AppDescription,
		Version: db.AppVersion,
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
	daeConsts.AppName = db.AppName

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(resetpassCmd)
}
