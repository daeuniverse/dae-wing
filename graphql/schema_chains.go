/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package graphql

import (
	"github.com/v2rayA/dae-wing/graphql/config"
	"github.com/v2rayA/dae-wing/graphql/config/dns"
	"github.com/v2rayA/dae-wing/graphql/config/global"
	"github.com/v2rayA/dae-wing/graphql/config/group"
	"github.com/v2rayA/dae-wing/graphql/config/routing"
	"github.com/v2rayA/dae-wing/graphql/service"
	"github.com/v2rayA/dae-wing/graphql/service/node"
	"github.com/v2rayA/dae-wing/graphql/service/subscription"
)

type SchemaChain func() (string, error)

var schemaChains = []SchemaChain{
	config.Schema,
	global.Schema,
	group.Schema,
	routing.Schema,
	dns.Schema,
	service.Schema,
	node.Schema,
	subscription.Schema,
}
