/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/db"
	daeConfig "github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/pkg/config_parser"
	"sort"
	"strings"
)

func Create(ctx context.Context, global string, dns string, routing string) (*Resolver, error) {
	m := db.Config{
		ID:       0,
		Global:   global,
		Dns:      dns,
		Routing:  routing,
		Selected: false,
	}
	c, err := m.ToDaeConfig()
	if err != nil {
		return nil, err
	}
	if err = db.DB(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return &Resolver{
		Config: c,
		Model:  &m,
	}, nil
}

func Update(ctx context.Context, _id graphql.ID, global *string, dns *string, routing *string) (*Resolver, error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return nil, err
	}
	tx := db.BeginTx(ctx)
	var m db.Config
	if err = tx.Model(&db.Config{}).Where("id = ?", id).First(&m).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	var updates map[string]interface{}
	if global != nil {
		m.Global = *global
		updates["global"] = *global
	}
	if dns != nil {
		m.Dns = *dns
		updates["dns"] = *dns
	}
	if routing != nil {
		m.Routing = *routing
		updates["routing"] = *routing
	}
	// Check grammar.
	c, err := m.ToDaeConfig()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = tx.Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &Resolver{
		Config: c,
		Model:  &m,
	}, nil
}

func Remove(ctx context.Context, _id graphql.ID) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	if err = db.DB(ctx).Delete(&db.Config{ID: id}).Error; err != nil {
		return 0, err
	}
	return 1, nil
}

func Select(ctx context.Context, _id graphql.ID) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	// Unset all selected.
	if err = tx.Model(&db.Config{}).Where("selected = ?", true).Update("selected", false).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	// Set selected.
	q := tx.Model(&db.Config{ID: id}).Update("selected", true)
	if err = q.Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if q.RowsAffected == 0 {
		tx.Rollback()
		return 0, fmt.Errorf("no such config")
	}
	tx.Commit()
	return 1, nil
}

func Run(ctx context.Context, dry bool) (n int32, err error) {
	// Dry run.
	if dry {
		// TODO: ...
		return 0, nil
	}

	// Run selected config.
	var m db.Config
	tx := db.DB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSnapshot,
		ReadOnly:  true,
	})
	defer tx.Commit()
	q := tx.Model(&db.Config{}).Where("selected = ?", true).First(&m)
	if q.Error != nil {
		return 0, q.Error
	}
	if q.RowsAffected == 0 {
		return 0, fmt.Errorf("please select a config")
	}
	c, err := m.ToDaeConfig()
	if err != nil {
		return 0, err
	}
	/// Fill in necessary groups and nodes.
	// Find groups needed by routing.
	outbounds := NecessaryOutbounds(&c.Routing)
	var groups []db.Group
	q = tx.Model(&db.Group{}).
		Where("name in ?", outbounds).
		Preload("PolicyParams").
		Preload("Subscription").
		Preload("Subscription.Node").
		Find(&groups)
	if q.Error != nil {
		return 0, q.Error
	}
	if q.RowsAffected != int64(len(outbounds)) {
		// Find not found.
		nameSet := map[string]struct{}{}
		for _, name := range outbounds {
			nameSet[name] = struct{}{}
		}
		for _, g := range groups {
			delete(nameSet, g.Name)
		}
		var notFound []string
		for name := range nameSet {
			notFound = append(notFound, name)
		}
		return 0, fmt.Errorf("groups not defined but referenced by routing: %v", strings.Join(notFound, ", "))
	}
	// Find nodes in groups.
	var nodes []db.Node
	var subIds []uint
	for _, g := range groups {
		for _, s := range g.Subscription {
			nodes = append(nodes, s.Node...)
			subIds = append(subIds, s.ID)
		}
	}
	var separateNodes []db.Node
	if err = tx.Model(&db.Group{}).
		Where("name in ?", outbounds).
		Association("Node").
		Find(separateNodes, "subscription_id not in ?", subIds); err != nil {
		return 0, err
	}
	nodes = append(nodes, separateNodes...)
	// Uniquely name nodes.
	// Sort nodes by "has node.Tag" because node.Tag is unique but names of others may be the same with them.
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Tag != nil && nodes[j].Tag == nil
	})
	var nameToNodes map[string]*db.Node
	for i := range nodes {
		node := &nodes[i]
		if node.Tag != nil {
			nameToNodes[*node.Tag] = node
		} else {
			baseName := node.Name
			if node.SubscriptionID != nil {
				baseName = fmt.Sprintf("%v.%v", *node.SubscriptionID, baseName)
			}
			// SubID.Name
			wantedName := baseName
			for j := 0; ; j++ {
				_, exist := nameToNodes[wantedName]
				if !exist {
					nameToNodes[wantedName] = node
					break
				}
				// SubID.Name.1
				wantedName = fmt.Sprintf("%v.%v", baseName, j)
			}
		}
	}
	// Fill in groups.
	for _, g := range groups {
		// Parse policy.
		var policy daeConfig.FunctionListOrString
		if len(g.PolicyParams) == 0 {
			policy = g.Policy
		} else {
			// Parse policy params.
			var params []*config_parser.Param
			for _, param := range g.PolicyParams {
				params = append(params, param.Marshal())
			}
			policy = &config_parser.Function{
				Name:   g.Policy,
				Not:    false,
				Params: params,
			}
		}
		// Node names to filter.
		var names []*config_parser.Param
		for name := range nameToNodes {
			names = append(names, &config_parser.Param{
				Val: name,
			})
		}
		c.Group = append(c.Group, daeConfig.Group{
			Name: g.Name,
			Filter: []*config_parser.Function{{
				Name:   "name",
				Not:    false,
				Params: names,
			}},
			Policy: policy,
		})
	}
	// Fill in nodes.
	for name, node := range nameToNodes {
		c.Node = append(c.Node, daeConfig.KeyableString(fmt.Sprintf("%v:%v", name, node.Link)))
	}

	/// Reload with current config.

	return 1, nil
}
