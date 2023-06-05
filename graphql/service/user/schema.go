/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package user

func Schema() (string, error) {
	return `
type User {
	username: String!
	name: String
	avatar: String
}
`, nil
}
