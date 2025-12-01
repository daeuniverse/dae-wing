/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"fmt"
	"strings"

	"github.com/graph-gophers/graphql-go"
)

var rootSchema = `
scalar Duration
scalar Time

directive @hasRole(role: Role!) on FIELD_DEFINITION

schema {
	query: Query
	mutation: Mutation
}
type Query {
	healthCheck: Int!
	token(username: String!, password: String!): String!
	numberUsers: Int!
	# jsonStorage get given paths from user related json storage. Empty paths is to get all. Refer to https://github.com/tidwall/gjson
	jsonStorage(paths: [String!]): [String!]! @hasRole(role: ADMIN)
    user: User! @hasRole(role: ADMIN)
	configFlatDesc: [ConfigFlatDesc!]! @hasRole(role: ADMIN)
	configs(id: ID, selected: Boolean): [Config!]! @hasRole(role: ADMIN)
	dnss(id: ID, selected: Boolean): [Dns!]! @hasRole(role: ADMIN)
	routings(id: ID, selected: Boolean): [Routing!]! @hasRole(role: ADMIN)
	parsedRouting(raw: String!): DaeRouting! @hasRole(role: ADMIN)
	parsedDns(raw: String!): DaeDns! @hasRole(role: ADMIN)
	subscriptions(id: ID): [Subscription!]! @hasRole(role: ADMIN)
	groups(id: ID): [Group!]! @hasRole(role: ADMIN)
	group(name: String!): Group! @hasRole(role: ADMIN)
	nodes(id: ID, subscriptionId: ID, first: Int, after: ID): NodesConnection! @hasRole(role: ADMIN)
	general: General! @hasRole(role: ADMIN)
}
type Mutation {
	# createUser creates a user if there is no user.
	createUser(username: String!, password: String!): String!
	# createConfig creates a global config. Null arguments will be converted to default value.
	createConfig(name: String, global: globalInput): Config! @hasRole(role: ADMIN)
	# createConfig creates a dns config. Null arguments will be converted to default value.
	createDns(name: String, dns: String): Dns! @hasRole(role: ADMIN)
	# createConfig creates a routing config. Null arguments will be converted to default value.
	createRouting(name: String, routing: String): Routing! @hasRole(role: ADMIN)

	# setJsonStorage set given paths to values in user related json storage. Refer to https://github.com/tidwall/sjson
	setJsonStorage(paths: [String!]!, values: [String!]!): Int! @hasRole(role: ADMIN)
	# removeJsonStorage remove given paths from user related json storage. Empty paths is to clear json storage. Refer to https://github.com/tidwall/sjson
	removeJsonStorage(paths: [String!]): Int! @hasRole(role: ADMIN)
	# updateAvatar update avatar for current user. Remove avatar if avatar is null. Blob base64 encoded image is recommended.
	updateAvatar(avatar: String): Int! @hasRole(role: ADMIN)
	# updateName update name for current user. Remove name if name is null.
	updateName(name: String): Int! @hasRole(role: ADMIN)
	# updateUsername update username for current user.
	updateUsername(username: String!): Int! @hasRole(role: ADMIN)
	# updatePassword update password for current user. currentPassword is needed to authenticate. Return new token.
	updatePassword(currentPassword: String!, newPassword: String!): String! @hasRole(role: ADMIN)

	# updateConfig allows to partially update global config with given id.
	updateConfig(id: ID!, global: globalInput!): Config! @hasRole(role: ADMIN)
	# updateDns is to update dns config with given id.
	updateDns(id: ID!, dns: String!): Dns! @hasRole(role: ADMIN)
	# updateRouting is to update routing config with given id.
	updateRouting(id: ID!, routing: String!): Routing! @hasRole(role: ADMIN)

	# renameConfig is to give the config a new name.
	renameConfig(id: ID!, name: String!): Int! @hasRole(role: ADMIN)
	# renameDns is to give the dns config a new name.
	renameDns(id: ID!, name: String!): Int! @hasRole(role: ADMIN)
	# renameRouting is to give the routing config a new name.
	renameRouting(id: ID!, name: String!): Int! @hasRole(role: ADMIN)

	# removeConfig is to remove a config with given config ID.
	removeConfig(id: ID!): Int! @hasRole(role: ADMIN)
	# removeDns is to remove a dns config with given dns ID.
	removeDns(id: ID!): Int! @hasRole(role: ADMIN)
	# removeRouting is to remove a routing config with given routing ID.
	removeRouting(id: ID!): Int! @hasRole(role: ADMIN)

	# selectConfig is to select a config as the current config.
	selectConfig(id: ID!): Int! @hasRole(role: ADMIN)
	# selectConfig is to select a dns config as the current dns.
	selectDns(id: ID!): Int! @hasRole(role: ADMIN)
	# selectConfig is to select a routing config as the current routing.
	selectRouting(id: ID!): Int! @hasRole(role: ADMIN)

	# run proxy with selected config+dns+routing. Dry-run can be used to stop the proxy.
	run(dry: Boolean!): Int! @hasRole(role: ADMIN)

	# importNodes is to import nodes with no subscription ID. rollbackError means abort the import on error.
	importNodes(rollbackError: Boolean!, args: [ImportArgument!]!): [NodeImportResult!]! @hasRole(role: ADMIN)

	# updateNode is to update a node with no subscription ID.
	updateNode(id: ID!, newLink: String!): Node! @hasRole(role: ADMIN)

	# removeNodes is to remove nodes that have no subscription ID.
	removeNodes(ids: [ID!]!): Int! @hasRole(role: ADMIN)

	# tagNode is to give the node a new tag.
	tagNode(id: ID!, tag: String!): Int! @hasRole(role: ADMIN)

	# importSubscription is to fetch and resolve the subscription into nodes.
	importSubscription(rollbackError: Boolean!, arg: ImportArgument!): SubscriptionImportResult! @hasRole(role: ADMIN)

	# removeSubscriptions is to remove subscriptions with given ID list.
	removeSubscriptions(ids: [ID!]!): Int! @hasRole(role: ADMIN)

	# tagSubscription is to give the subscription a new tag.
	tagSubscription(id: ID!, tag: String!): Int! @hasRole(role: ADMIN)

	# updateSubscription is to re-fetch subscription and resolve subscription into nodes. Old nodes that independently belong to any groups will not be removed.
	updateSubscription(id: ID!): Subscription! @hasRole(role: ADMIN)

	# updateSubscriptionLink is to update the subscription link without re-fetching nodes.
	updateSubscriptionLink(id: ID!, link: String!): Subscription! @hasRole(role: ADMIN)

	# createGroup is to create a group.
	createGroup(name: String!, policy: Policy!, policyParams: [PolicyParam!]): Group! @hasRole(role: ADMIN)

	# groupSetPolicy is to set the group a new policy.
	groupSetPolicy(id: ID!, policy: Policy!, policyParams: [PolicyParam!]): Int! @hasRole(role: ADMIN)

	# groupAddSubscriptions is to add subscriptions to the group.
	groupAddSubscriptions(id: ID!, subscriptionIDs: [ID!]!): Int! @hasRole(role: ADMIN)

	# groupDelSubscriptions is to remove subscriptions from the group.
	groupDelSubscriptions(id: ID!, subscriptionIDs: [ID!]!): Int! @hasRole(role: ADMIN)

	# groupAddNodes is to add nodes to the group. Nodes will not be removed from its subscription when subscription update.
	groupAddNodes(id: ID!, nodeIDs: [ID!]!): Int! @hasRole(role: ADMIN)

	# groupDelNodes is to remove nodes from the group.
	groupDelNodes(id: ID!, nodeIDs: [ID!]!): Int! @hasRole(role: ADMIN)

	# renameGroup is to rename a group.
	renameGroup(id: ID!, name: String!): Int! @hasRole(role: ADMIN)

	# removeGroup is to remove a group.
	removeGroup(id: ID!): Int! @hasRole(role: ADMIN)
}
enum Role {
	ADMIN
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
	key: String
	val: String!
}
type ConfigFlatDesc {
	name: String!
	mapping: String!
	isArray: Boolean!
	defaultValue: String!
	required: Boolean!
	type: String!
	desc: String!
}
`

type hasRoleDirective struct {
	Role string
}

func (h *hasRoleDirective) ImplementsDirective() string {
	return "hasRole"
}

func (h *hasRoleDirective) Validate(ctx context.Context, _ interface{}) error {
	role := ctx.Value("role")
	if role == nil {
		return fmt.Errorf("access denied")
	}
	if !strings.EqualFold(role.(string), h.Role) {
		return fmt.Errorf("access denied, %q role required", h.Role)
	}
	return nil
}

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
		graphql.Directives(&hasRoleDirective{}),
	), nil
}
