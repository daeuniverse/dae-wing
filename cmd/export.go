/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/v2rayA/dae-wing/graphql"
	"os"
)

var (
	exportCmd = &cobra.Command{
		Use: "export",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	exportOutlineCmd = &cobra.Command{
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
)

func init() {
	exportCmd.AddCommand(exportOutlineCmd)
}
