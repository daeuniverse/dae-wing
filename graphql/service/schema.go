/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package service

func Schema() (string, error) {
	return `
type PageInfo {
	startCursor: ID
	endCursor: ID
	hasNextPage: Boolean!
}
`, nil
}
