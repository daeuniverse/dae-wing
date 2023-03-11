/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"errors"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/service/config"
	"github.com/daeuniverse/dae-wing/graphql/service/dns"
	"github.com/daeuniverse/dae-wing/graphql/service/general"
	"github.com/daeuniverse/dae-wing/graphql/service/group"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/daeuniverse/dae-wing/graphql/service/routing"
	"github.com/daeuniverse/dae-wing/graphql/service/subscription"
	"github.com/graph-gophers/graphql-go"
	daeConfig "github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/pkg/config_parser"
	"gorm.io/gorm"
)

type queryResolver struct{}

func (r *queryResolver) General() *general.Resolver {
	return &general.Resolver{}
}
func (r *queryResolver) HealthCheck() int32 {
	return 1
}
func (r *queryResolver) Configs(args *struct {
	ID       *graphql.ID
	Selected *bool
}) (rs []*config.Resolver, err error) {
	// Check if query specific ID.
	var id uint
	q := db.DB(context.TODO()).Model(&db.Config{})
	if args.ID != nil {
		id, err = common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	if args.Selected != nil {
		q = q.Where("selected = ?", *args.Selected)
	}
	// Get configs from DB.
	var models []db.Config
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	for i := range models {
		m := &models[i]
		c, err := dae.ParseConfig(&m.Global, nil, nil)
		if err != nil {
			return nil, err
		}
		rs = append(rs, &config.Resolver{
			DaeGlobal: &c.Global,
			Model:     m,
		})
	}
	return rs, nil
}

func (r *queryResolver) Dnss(args *struct {
	ID       *graphql.ID
	Selected *bool
}) (rs []*dns.Resolver, err error) {
	// Check if query specific ID.
	var id uint
	q := db.DB(context.TODO()).Model(&db.Dns{})
	if args.ID != nil {
		id, err = common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	if args.Selected != nil {
		q = q.Where("selected = ?", *args.Selected)
	}
	// Get dns from DB.
	var models []db.Dns
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	for i := range models {
		m := &models[i]
		c, err := dae.ParseConfig(nil, &m.Dns, nil)
		if err != nil {
			return nil, err
		}
		rs = append(rs, &dns.Resolver{
			DaeDns: &c.Dns,
			Model:  m,
		})
	}
	return rs, nil
}

func (r *queryResolver) Routings(args *struct {
	ID       *graphql.ID
	Selected *bool
}) (rs []*routing.Resolver, err error) {
	// Check if query specific ID.
	var id uint
	q := db.DB(context.TODO()).Model(&db.Routing{})
	if args.ID != nil {
		id, err = common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	if args.Selected != nil {
		q = q.Where("selected = ?", *args.Selected)
	}
	// Get routing from DB.
	var models []db.Routing
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	for i := range models {
		m := &models[i]
		c, err := dae.ParseConfig(nil, nil, &m.Routing)
		if err != nil {
			return nil, err
		}
		rs = append(rs, &routing.Resolver{
			DaeRouting: &c.Routing,
			Model:      m,
		})
	}
	return rs, nil
}

func (r *queryResolver) ConfigFlatDesc() []*dae.FlatDesc {
	return dae.ExportFlatDesc()
}
func (r *queryResolver) ParsedRouting(args *struct{ Raw string }) (rr *routing.DaeResolver, err error) {
	sections, err := config_parser.Parse("global{} routing {" + args.Raw + "}")
	if err != nil {
		return nil, err
	}
	conf, err := daeConfig.New(sections)
	if err != nil {
		return nil, err
	}
	return &routing.DaeResolver{
		Routing: &conf.Routing,
	}, nil
}
func (r *queryResolver) ParsedDns(args *struct{ Raw string }) (dr *dns.DaeResolver, err error) {
	sections, err := config_parser.Parse("global{} dns {" + args.Raw + "} routing{}")
	if err != nil {
		return nil, err
	}
	conf, err := daeConfig.New(sections)
	if err != nil {
		return nil, err
	}
	return &dns.DaeResolver{
		Dns: &conf.Dns,
	}, nil
}
func (r *queryResolver) Subscriptions(args *struct{ ID *graphql.ID }) (rs []*subscription.Resolver, err error) {
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

func (r *queryResolver) Group(args *struct{ Name string }) (rs *group.Resolver, err error) {
	var m db.Group
	if err = db.DB(context.TODO()).
		Model(&db.Group{}).
		Where("name = ?", args.Name).First(&m).Error; err != nil {
		return nil, err
	}
	return &group.Resolver{Group: &m}, nil
}
func (r *queryResolver) Groups(args *struct{ ID *graphql.ID }) (rs []*group.Resolver, err error) {
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
func (r *queryResolver) Nodes(args *struct {
	ID             *graphql.ID
	SubscriptionID *graphql.ID
	First          *int32
	After          *graphql.ID
}) (rs *node.ConnectionResolver, err error) {
	return node.NewConnectionResolver(args.ID, args.SubscriptionID, args.First, args.After)
}
