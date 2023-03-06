/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package node

func Schema() (string, error) {
	return `
type Node {
	id: ID!
	link: String!
	name: String!
	address: String!
	protocol: String!
	tag: String
	subscriptionID: ID
}
type NodesConnection {
	totalCount: Int!
	edges: [Node!]!
	pageInfo: PageInfo! 
}
`, nil
}
