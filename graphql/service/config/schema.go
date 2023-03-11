/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package config

func Schema() (string, error) {
	return `
type Config {
	id: ID!
	name: String!
	global: Global!
	selected: Boolean!
}
`, nil
}
