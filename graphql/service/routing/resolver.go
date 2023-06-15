/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package routing

import (
	"strings"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/daeuniverse/dae/pkg/config_parser"
	"github.com/graph-gophers/graphql-go"
)

type Resolver struct {
	DaeRouting *daeConfig.Routing
	Model      *db.Routing
}

func (r *Resolver) ID() graphql.ID {
	return common.EncodeCursor(r.Model.ID)
}

func (r *Resolver) Name() string {
	return r.Model.Name
}

func (r *Resolver) Routing() *DaeResolver {
	return &DaeResolver{
		Routing: r.DaeRouting,
		Raw:     r.Model.Routing,
	}
}

func (r *Resolver) Selected() bool {
	return r.Model.Selected
}

func (r *Resolver) ReferenceGroups() (outbounds []string) {
	return dae.NecessaryOutbounds(r.DaeRouting)
}

type DaeResolver struct {
	*daeConfig.Routing
	Raw string
}

func (r *DaeResolver) String() (string, error) {
	section := strings.TrimSpace(r.Raw)
	section = strings.TrimPrefix(section, "routing {")
	section = strings.TrimSuffix(section, "}")
	return strings.TrimSpace(section), nil
}

func (r *DaeResolver) Rules() (rs []*RuleResolver) {
	for _, rule := range r.Routing.Rules {
		rs = append(rs, &RuleResolver{RoutingRule: rule})
	}
	return rs
}

func (r *DaeResolver) Fallback() *internal.FunctionOrPlaintextResolver {
	return &internal.FunctionOrPlaintextResolver{FunctionOrString: r.Routing.Fallback}
}

type RuleResolver struct {
	*config_parser.RoutingRule
}

func (r *RuleResolver) Conditions() *internal.AndFunctionsResolver {
	return &internal.AndFunctionsResolver{AndFunctions: r.AndFunctions}
}

func (r *RuleResolver) Outbound() *internal.FunctionResolver {
	return &internal.FunctionResolver{Function: &r.RoutingRule.Outbound}
}
