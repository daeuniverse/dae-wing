/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package general

func Schema() (string, error) {
	return `
type General {
  dae: Dae!
  interfaces(up:Boolean): [Interface!]!
}
type Dae {
  running: Boolean!
}
type Interface {
  name: String!
  flag: InterfaceFlag!
  ifindex: Int!
  ip(onlyGlobalScope:Boolean): [String!]!
}
type InterfaceFlag {
  up: Boolean!
  default: [DefaultRoute!]
}
type DefaultRoute {
  ipVersion: String!
  gateway: String
  source: String
}
`, nil
}
