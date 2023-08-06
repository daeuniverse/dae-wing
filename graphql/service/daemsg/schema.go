/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package daemsg

func Schema() (string, error) {
	return `
type DaeMsg {
	type: MsgType!
	timestamp: Time!
	checkResult: CheckResult
}
enum MsgType {
	connectivityCheckDone
}
type CheckResult {
	dialerProperty: Property!
	checkType: NetworkType!
	latency: Int64!
	alive: Boolean!
	error: String
}
type Property {
	name: String!
	address: String!
	protocol: String!
	link: String!
}
type NetworkType {
	l4Proto: String!
	ipVersion: String!
	isDns: Boolean!
}
`, nil
}
