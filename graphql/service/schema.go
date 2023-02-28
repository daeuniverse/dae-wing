/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package service

func Schema() (string, error) {
	return `
type Model {
	id: ID!
	created_at: Time!
	updated_at: Time!
	deleted_at: Time
}
type PageInfo {
	start_cursor: ID
	end_cursor: ID
	has_next_page: Boolean!
}
`, nil
}
