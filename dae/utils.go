/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dae

import (
	daeCommon "github.com/v2rayA/dae/common"
	daeConfig "github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/pkg/config_parser"
	"strings"
)

var (
	EmptyGroupSection        = `group {}`
	EmptySubscriptionSection = `subscription {}`
	EmptyNodeSection         = `node {}`
	EmptyRoutingSection      = `routing {}`
	EmptyDnsSection          = `dns {}`
	EmptyGlobalSection       = `global {}`
)

func NecessaryOutbounds(routing *daeConfig.Routing) (outbounds []string) {
	f := daeConfig.FunctionOrStringToFunction(routing.Fallback)
	outbounds = append(outbounds, f.Name)
	for _, r := range routing.Rules {
		outbounds = append(outbounds, r.Outbound.Name)
	}
	return daeCommon.Deduplicate(outbounds)
}

func ParseConfig(globalSection *string, dnsSection *string, routingSection *string) (*daeConfig.Config, error) {
	if globalSection == nil {
		globalSection = &EmptyGlobalSection
	}
	if dnsSection == nil {
		dnsSection = &EmptyDnsSection
	}
	if routingSection == nil {
		routingSection = &EmptyRoutingSection
	}
	strConfig := strings.Join([]string{
		*globalSection,
		*dnsSection,
		*routingSection,
		EmptyGroupSection,
		EmptySubscriptionSection,
		EmptyNodeSection,
	}, "\n")
	// Parse it to sections.
	sections, err := config_parser.Parse(strConfig)
	if err != nil {
		return nil, err
	}
	// New dae.Config from sections.
	c, err := daeConfig.New(sections)
	if err != nil {
		return nil, err
	}
	return c, err
}
