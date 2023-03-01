/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package group

func Schema() (string, error) {
	return `
type Group {
	name: String!
	nodes: [Node!]!
	subscriptions: [Subscription!]!
	policy: AndFunctionsOrPlaintext!
}
`, nil
}
