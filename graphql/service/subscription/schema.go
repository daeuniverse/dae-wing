/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package subscription

func Schema() (string, error) {
	return `
type Subscription {
	model: Model!
	remarks: String
	link: String!
	status: String!
	info: String!
	nodes(first: Int, after: ID): NodesConnection!
}
`, nil
}
