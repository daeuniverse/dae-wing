/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/v2rayA/dae-wing/graphql"
	daeConfig "github.com/v2rayA/dae/config"
	"os"
)

var (
	exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export development related information",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	exportSchemaCmd = &cobra.Command{
		Use: "schema",
		Run: func(cmd *cobra.Command, args []string) {
			schema, err := graphql.SchemaString()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(schema)
		},
	}
	exportOutlineCmd = &cobra.Command{
		Use: "outline",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(daeConfig.ExportOutlineJson(Version))
		},
	}
)

func init() {
	exportCmd.AddCommand(exportSchemaCmd)
	exportCmd.AddCommand(exportOutlineCmd)
}
