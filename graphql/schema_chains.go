/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package graphql

import (
	"github.com/daeuniverse/dae-wing/graphql/service"
	"github.com/daeuniverse/dae-wing/graphql/service/config"
	"github.com/daeuniverse/dae-wing/graphql/service/config/global"
	"github.com/daeuniverse/dae-wing/graphql/service/daemsg"
	"github.com/daeuniverse/dae-wing/graphql/service/dns"
	"github.com/daeuniverse/dae-wing/graphql/service/general"
	"github.com/daeuniverse/dae-wing/graphql/service/group"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/daeuniverse/dae-wing/graphql/service/routing"
	"github.com/daeuniverse/dae-wing/graphql/service/subscription"
	"github.com/daeuniverse/dae-wing/graphql/service/user"
)

type SchemaChain func() (string, error)

var schemaChains = []SchemaChain{
	general.Schema,
	config.Schema,
	global.Schema,
	group.Schema,
	routing.Schema,
	dns.Schema,
	service.Schema,
	node.Schema,
	subscription.Schema,
	user.Schema,
	daemsg.Schema,
}
