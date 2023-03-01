/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package routing

import (
	"github.com/v2rayA/dae-wing/graphql/config"
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

func (r *Resolver) Fallback() *config.FunctionOrPlaintextResolver {
	return &config.FunctionOrPlaintextResolver{FunctionOrString: r.Routing.Fallback}
}

type RuleResolver struct {
	*config_parser.RoutingRule
}

func (r *RuleResolver) Conditions() *config.AndFunctionsResolver {
	return &config.AndFunctionsResolver{AndFunctions: r.AndFunctions}
}

func (r *RuleResolver) Outbound() *config.FunctionResolver {
	return &config.FunctionResolver{Function: &r.RoutingRule.Outbound}
}
