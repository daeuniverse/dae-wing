/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package config

func Schema() (string, error) {
	return `
type Config {
	global: Global!
}
`, nil
}
