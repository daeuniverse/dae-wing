/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dns

import (
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	"github.com/daeuniverse/dae-wing/graphql/service/routing"
	daeCommon "github.com/daeuniverse/dae/common"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/daeuniverse/dae/pkg/config_parser"
	"github.com/graph-gophers/graphql-go"
	"reflect"
	"strings"
)

type Resolver struct {
	DaeDns *daeConfig.Dns
	Model  *db.Dns
}

func (r *Resolver) ID() graphql.ID {
	return common.EncodeCursor(r.Model.ID)
}

func (r *Resolver) Name() string {
	return r.Model.Name
}

func (r *Resolver) Dns() *DaeResolver {
	return &DaeResolver{
		Dns: r.DaeDns,
	}
}

func (r *Resolver) Selected() bool {
	return r.Model.Selected
}

type DaeResolver struct {
	*daeConfig.Dns
}

func (r *DaeResolver) String() (string, error) {
	marshaller := daeConfig.Marshaller{IndentSpace: 2}
	if err := marshaller.MarshalSection("dns", reflect.ValueOf(*r.Dns), -1); err != nil {
		return "", err
	}
	section := strings.TrimSpace(string(marshaller.Bytes()))
	section = strings.TrimPrefix(section, "dns {")
	section = strings.TrimSuffix(section, "}")
	return strings.TrimSpace(section), nil
}

func (r *DaeResolver) Upstream() (rs []*internal.ParamResolver) {
	for _, upstream := range r.Dns.Upstream {
		tag, afterTag := daeCommon.GetTagFromLinkLikePlaintext(string(upstream))
		rs = append(rs, &internal.ParamResolver{Param: &config_parser.Param{
			Key: tag,
			Val: afterTag,
		}})
	}
	return rs
}

func (r *DaeResolver) Routing() *RoutingResolver {
	return &RoutingResolver{DnsRouting: &r.Dns.Routing}
}

type RoutingResolver struct {
	*daeConfig.DnsRouting
}

func (r *RoutingResolver) Request() *routing.DaeResolver {
	return &routing.DaeResolver{Routing: &daeConfig.Routing{
		Rules:    r.DnsRouting.Request.Rules,
		Fallback: r.DnsRouting.Request.Fallback,
	}}
}
func (r *RoutingResolver) Response() *routing.DaeResolver {
	return &routing.DaeResolver{Routing: &daeConfig.Routing{
		Rules:    r.DnsRouting.Response.Rules,
		Fallback: r.DnsRouting.Response.Fallback,
	}}
}
