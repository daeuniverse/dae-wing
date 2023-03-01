/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package dns

func Schema() (string, error) {
	return `
type Dns {
	upstream: [Param!]!
	routing: DnsRouting!
}
type DnsRouting {
	request: Routing!
	response: Routing!
}
`, nil
}
