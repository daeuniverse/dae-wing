/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package routing

func Schema() (string, error) {
	return `
type Routing {
	rules: [RoutingRule!]!
	fallback: FunctionOrPlaintext!
}
type RoutingRule {
	conditions: AndFunctions!
	outbound: Function!
}
`, nil
}
