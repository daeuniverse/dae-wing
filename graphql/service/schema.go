/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package service

func Schema() (string, error) {
	return `
type Model {
	id: ID!
	createdAt: Time!
	updatedAt: Time!
	deletedAt: Time
}
type PageInfo {
	startCursor: ID
	endCursor: ID
	hasNextPage: Boolean!
}
`, nil
}
