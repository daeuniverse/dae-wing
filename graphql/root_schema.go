/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/config/global"
	"github.com/v2rayA/dae-wing/graphql/service/subscription"
	"github.com/v2rayA/dae-wing/model"
	"github.com/v2rayA/dae/config"
	"strings"
)

var rootSchema = `
scalar Duration
scalar Time
//scalar Int8, Int16, Int32, Int64
//scalar UInt8, UInt16, UInt32, UInt64

schema {
	query: Query
	//mutation: Mutation
}
type Query {
	config: Config!
	subscriptions(id: Int): [Subscription!]!
}
`

type Resolver struct{}

func (*Resolver) Query() *QueryResolver {
	return &QueryResolver{}
}

type QueryResolver struct{}

func (r *QueryResolver) Config() (*configResolver, error) {
	m := config.NewMerger("/home/mzz/ebpfProjects/ragdoll/foo/example.dae")
	sections, _, err := m.Merge()
	if err != nil {
		return nil, err
	}
	c, err := config.New(sections)
	if err != nil {
		return nil, err
	}
	return &configResolver{
		Config: c,
	}, nil
}

func (r *QueryResolver) Subscriptions(args struct{ ID graphql.NullInt }) (rs []*subscription.Resolver, err error) {
	q := db.DB(context.TODO()).
		Model(&model.SubscriptionModel{})
	if args.ID.Set {
		q = q.Where("id == ?", *args.ID.Value)
	}
	var models []model.SubscriptionModel
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	for _, _m := range models {
		m := _m
		rs = append(rs, &subscription.Resolver{
			SubscriptionModel: &m,
		})
	}
	return rs, nil
}

type configResolver struct {
	*config.Config
}

func (r *configResolver) Global() *global.Resolver {
	return &global.Resolver{
		Global: &r.Config.Global,
	}
}

func Schema() (*graphql.Schema, error) {
	var sb strings.Builder
	sb.WriteString(rootSchema)
	for _, c := range schemaChains {
		s, err := c()
		if err != nil {
			return nil, err
		}
		sb.WriteString(s)
	}
	return graphql.MustParseSchema(sb.String(), &QueryResolver{}), nil
}
