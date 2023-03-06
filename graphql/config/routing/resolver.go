/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package routing

import (
	"github.com/daeuniverse/dae-wing/graphql/internal"
	daeConfig "github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/pkg/config_parser"
)

type Resolver struct {
	*daeConfig.Routing
}

func (r *Resolver) Rules() (rs []*RuleResolver) {
	for _, rule := range r.Routing.Rules {
		rs = append(rs, &RuleResolver{RoutingRule: rule})
	}
	return rs
}

func (r *Resolver) Fallback() *internal.FunctionOrPlaintextResolver {
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
