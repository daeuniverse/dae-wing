/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package graphql

import (
	"github.com/daeuniverse/dae-wing/graphql/config"
	"github.com/daeuniverse/dae-wing/graphql/config/dns"
	"github.com/daeuniverse/dae-wing/graphql/config/global"
	"github.com/daeuniverse/dae-wing/graphql/config/routing"
	"github.com/daeuniverse/dae-wing/graphql/general"
	"github.com/daeuniverse/dae-wing/graphql/service"
	"github.com/daeuniverse/dae-wing/graphql/service/group"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/daeuniverse/dae-wing/graphql/service/subscription"
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
	general.Schema,
}
