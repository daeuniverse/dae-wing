/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package cmd

import (
	"fmt"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/graphql"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
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
	exportFlatDescCmd = &cobra.Command{
		Use: "flatdesc",
		Run: func(cmd *cobra.Command, args []string) {
			b, _ := jsoniter.MarshalIndent(map[string]interface{}{
				"Version": Version,
				"Desc":    dae.ExportFlatDesc(),
			}, "", "  ")
			fmt.Println(string(b))
		},
	}
)

func init() {
	exportCmd.AddCommand(exportSchemaCmd)
	exportCmd.AddCommand(exportOutlineCmd)
	exportCmd.AddCommand(exportFlatDescCmd)
}
