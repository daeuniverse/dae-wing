/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package config

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/config/dns"
	"github.com/v2rayA/dae-wing/graphql/config/global"
	"github.com/v2rayA/dae-wing/graphql/config/routing"
	"github.com/v2rayA/dae/config"
)

type Resolver struct {
	*config.Config
	Model *db.Config
}

func (r *Resolver) ID() graphql.ID {
	return common.EncodeCursor(r.Model.ID)
}
func (r *Resolver) Global() *global.Resolver {
	return &global.Resolver{
		Global: &r.Config.Global,
	}
}

func (r *Resolver) Routing() *routing.Resolver {
	return &routing.Resolver{
		Routing: &r.Config.Routing,
	}
}

func (r *Resolver) Dns() *dns.Resolver {
	return &dns.Resolver{
		Dns: &r.Config.Dns,
	}
}

func (r *Resolver) Selected() bool {
	return r.Model.Selected
}
