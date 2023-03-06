/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package group

func Schema() (string, error) {
	return `
type Group {
	id: ID!
	name: String!
	nodes: [Node!]!
	subscriptions: [Subscription!]!
	policy: Policy!
	policyParams: [Param!]!
}
enum Policy {
	random
	fixed
	min_avg10
	min_moving_avg
	min
}
`, nil
}
