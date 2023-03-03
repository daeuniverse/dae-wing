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
	healthCheck: Int!
	configs(id: ID, selected: Boolean): [Config!]!
	subscriptions(id: ID): [Subscription!]!
	groups(id: ID): [Group!]!
	group(name: String!): Group!
	nodes(id: ID, subscriptionId: ID, first: Int, after: ID): NodesConnection!
}
type Mutation {
	# createConfig create a config. Null arguments will be converted to default value.
	createConfig(global: globalInput, dns: String, routing: String): Config!

	# updateConfig allows to partially update "global".
	updateConfig(id: ID!, global: globalInput, dns: String, routing: String): Config!

	# removeConfig is to remove a config with given config ID.
	removeConfig(id: ID!): Int!

	# selectConfig is to select a config as the current config.
	selectConfig(id: ID!): Int!

	# run proxy with the current config. Dry-run can be used to stop the proxy.
	run(dry: Boolean!): Int!

	# importNodes is to import nodes with no subscription ID. rollbackError means abort the import on error.
	importNodes(rollbackError: Boolean!, args: [ImportArgument!]!): [NodeImportResult!]!

	# removeNodes is to remove nodes that have no subscription ID.
	removeNodes(ids: [ID!]!): Int!

	# tagNode is to give the node a tag.
	tagNode(id: ID!, tag: String!): Int!

	# importSubscription is to fetch and resolve the subscription into nodes.
	importSubscription(rollbackError: Boolean!, arg: ImportArgument!): SubscriptionImportResult!

	# removeSubscriptions is to remove subscriptions with given ID list.
	removeSubscriptions(ids: [ID!]!): Int!

	# tagNode is to give the subscription a tag.
	tagSubscription(id: ID!, tag: String!): Int!

	# updateSubscription is to re-fetch subscription and resolve subscription into nodes. Old nodes that independently belong to any groups will not be removed.
	updateSubscription(id: ID!): Subscription!

	# createGroup is to create a group.
	createGroup(name: String!, policy: Policy!, policyParams: [PolicyParam!]): Group!

	# groupAddSubscriptions is to add subscriptions to the group.
	groupAddSubscriptions(id: ID!, subscriptionIDs: [ID!]!): Int!

	# groupDelSubscriptions is to remove subscriptions from the group.
	groupDelSubscriptions(id: ID!, subscriptionIDs: [ID!]!): Int!

	# groupAddNodes is to add nodes to the group. Nodes will not be removed from its subscription when subscription update.
	groupAddNodes(id: ID!, nodeIDs: [ID!]!): Int!

	# groupDelNodes is to remove nodes from the group.
	groupDelNodes(id: ID!, nodeIDs: [ID!]!): Int!

	# renameGroup is to rename a group.
	renameGroup(id: ID!, name: String!): Int!

	# removeGroup is to remove a group.
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
	nodeImportResult: [NodeImportResult!]!
	sub: Subscription!
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
