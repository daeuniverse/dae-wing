/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dns

func Schema() (string, error) {
	return `
type Dns {
	id: ID!
	name: String!
	dns: DaeDns!
	selected: Boolean!
}
type DaeDns {
	string: String!
	upstream: [Param!]!
	routing: DnsRouting!
}
type DnsRouting {
	request: DaeRouting!
	response: DaeRouting!
}
`, nil
}
