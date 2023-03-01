/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"strings"
)

var rootSchema = `
scalar Duration
scalar Time

schema {
	query: Query
	mutation: Mutation
}
type Query {
	config: Config!
	subscriptions(id: ID): [Subscription!]!
	groups(id: ID): [Group!]!
	group(name: String!): Group
	nodes(id: ID, subscriptionId: ID, first: Int, after: ID): NodesConnection!
}
type Mutation {
	importNodes(rollbackError: Boolean!, args: [ImportArgument!]!): [NodeImportResult!]!
	removeNodes(ids: [ID!]!): Int!
	tagNode(id: ID!, tag: String!): Int!

	importSubscription(rollbackError: Boolean!, arg: ImportArgument!): SubscriptionImportResult!
	removeSubscriptions(ids: [ID!]!): Int!
	tagSubscription(id: ID!, tag: String!): Int!

	createGroup(name: String!, policy: Policy!, policyParams: [PolicyParam!]): Group!
	groupAddSubscriptions(id: ID!, subscriptionIDs: [ID!]!): Int!
	groupDelSubscriptions(id: ID!, subscriptionIDs: [ID!]!): Int!
	groupAddNodes(id: ID!, nodeIDs: [ID!]!): Int!
	groupDelNodes(id: ID!, nodeIDs: [ID!]!): Int!
	renameGroup(id: ID!, name: String!): Int!
	removeGroup(id: ID!): Int!
}
input ImportArgument {
	link: String!
	tag: String
}
type NodeImportResult {
	link: String!
	error: String
	node: Node
}
type SubscriptionImportResult {
	link: String!
	error: String
	subscription: Subscription
}
input PolicyParam {
	key: String!
	val: String!
}
`

type resolver struct{}

func (*resolver) Query() *queryResolver {
	return &queryResolver{}
}

func (*resolver) Mutation() *MutationResolver {
	return &MutationResolver{}
}

func SchemaString() (string, error) {
	var sb strings.Builder
	sb.WriteString(rootSchema)
	for _, c := range schemaChains {
		s, err := c()
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
	}
	return strings.TrimSpace(sb.String()), nil
}

func Schema() (*graphql.Schema, error) {
	schema, err := SchemaString()
	if err != nil {
		return nil, err
	}
	return graphql.MustParseSchema(
		schema,
		&resolver{},
		graphql.UseFieldResolvers(),
		graphql.Directives(),
	), nil
}
