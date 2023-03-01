/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/config/dns"
	"github.com/v2rayA/dae-wing/graphql/config/global"
	"github.com/v2rayA/dae-wing/graphql/config/routing"
	"github.com/v2rayA/dae-wing/graphql/service/group"
	"github.com/v2rayA/dae-wing/graphql/service/node"
	"github.com/v2rayA/dae-wing/graphql/service/subscription"
	"github.com/v2rayA/dae/config"
	"gorm.io/gorm"
)

type queryResolver struct{}

func (r *queryResolver) Config() (*configResolver, error) {
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

func (r *queryResolver) Subscriptions(args struct{ ID *graphql.ID }) (rs []*subscription.Resolver, err error) {
	q := db.DB(context.TODO()).
		Model(&db.Subscription{})
	if args.ID != nil {
		id, err := common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	var models []db.Subscription
	if err = q.Find(&models).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	for _, _m := range models {
		m := _m
		rs = append(rs, &subscription.Resolver{
			Subscription: &m,
		})
	}
	return rs, nil
}

func (r *queryResolver) Group(args struct{ Name string }) (rs *group.Resolver, err error) {
	var m db.Group
	if err = db.DB(context.TODO()).
		Model(&db.Group{}).
		Where("name = ?", args.Name).First(&m).Error; err != nil {
		return nil, err
	}
	return &group.Resolver{Group: &m}, nil
}
func (r *queryResolver) Groups(args struct{ ID *graphql.ID }) (rs []*group.Resolver, err error) {
	q := db.DB(context.TODO()).
		Model(&db.Group{})
	if args.ID != nil {
		id, err := common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	var models []db.Group
	if err = q.Find(&models).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	for _, _m := range models {
		m := _m
		rs = append(rs, &group.Resolver{
			Group: &m,
		})
	}
	return rs, nil
}
func (r *queryResolver) Nodes(args struct {
	ID             *graphql.ID
	SubscriptionID *graphql.ID
	First          *int32
	After          *graphql.ID
}) (rs *node.ConnectionResolver, err error) {
	return node.NewConnectionResolver(args.ID, args.SubscriptionID, args.First, args.After)
}

type configResolver struct {
	*config.Config
}

func (r *configResolver) Global() *global.Resolver {
	return &global.Resolver{
		Global: &r.Config.Global,
	}
}

func (r *configResolver) Routing() *routing.Resolver {
	return &routing.Resolver{
		Routing: &r.Config.Routing,
	}
}

func (r *configResolver) Dns() *dns.Resolver {
	return &dns.Resolver{
		Dns: &r.Config.Dns,
	}
}
