/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package node

func Schema() (string, error) {
	return `
type Node {
	model: Model!
	link: String!
	name: String!
	protocol: String!
	remarks: String!
	status: String!
	subscription_id: ID
}
type NodesConnection {
	total_count: Int!
	edges: [Node!]!
	page_info: PageInfo! 
}
`, nil
}
