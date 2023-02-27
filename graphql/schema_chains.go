/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package graphql

import "github.com/v2rayA/dae-wing/graphql/config/global"

type SchemaChain func() (string, error)

var schemaChains = []SchemaChain{
	global.SubSchema,
}
