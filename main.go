/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package main

import (
	"github.com/daeuniverse/dae-wing/cmd"
	"github.com/json-iterator/go/extra"
	"os"
)

import (
	_ "github.com/daeuniverse/dae/component/outbound"
)

func main() {
	extra.RegisterFuzzyDecoders()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
