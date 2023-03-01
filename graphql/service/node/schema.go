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
	address: String!
	protocol: String!
	remarks: String!
	subscriptionID: ID
}
type NodesConnection {
	totalCount: Int!
	edges: [Node!]!
	pageInfo: PageInfo! 
}
`, nil
}
