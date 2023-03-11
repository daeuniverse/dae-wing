/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package routing

func Schema() (string, error) {
	return `
type Routing {
	id: ID!
	name: String!
	routing: DaeRouting!
	selected: Boolean!
	referenceGroups: [String!]!
}
type DaeRouting {
	string: String!
	rules: [RoutingRule!]!
	fallback: FunctionOrPlaintext!
}
type RoutingRule {
	conditions: AndFunctions!
	outbound: Function!
}
type Function {
	name: String!
	not: Boolean!
	params: [Param!]!
}
type Param {
	key: String!
	val: String!
}

type AndFunctions {
	and: [Function!]!
}

type Plaintext {
	val: String!
}

union AndFunctionsOrPlaintext = AndFunctions | Plaintext
union FunctionOrPlaintext = Function | Plaintext
`, nil
}
