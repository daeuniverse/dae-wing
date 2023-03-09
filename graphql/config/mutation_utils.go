/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package config

import (
	"context"
	"fmt"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/config/global"
	"github.com/graph-gophers/graphql-go"
	daeConfig "github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/pkg/config_parser"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
	"sort"
	"strings"
)

func Create(ctx context.Context, glob *global.Input, dns string, routing string) (*Resolver, error) {
	if glob == nil {
		glob = &global.Input{}
	}
	strGlobal, err := glob.Marshal()
	if err != nil {
		return nil, err
	}
	if dns == "" {
		dns = "dns {}"
	}
	if routing == "" {
		routing = "routing {}"
	}
	m := db.Config{
		ID:       0,
		Global:   strGlobal,
		Dns:      dns,
		Routing:  routing,
		Selected: false,
	}
	// Check grammar and to dae config.
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

func Update(ctx context.Context, _id graphql.ID, inputGlobal *global.Input, dns *string, routing *string) (*Resolver, error) {
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
	updates := map[string]interface{}{}
	if inputGlobal != nil {
		// Convert global string in database to daeConfig.Global.
		c, err := m.ToDaeConfig()
		if err != nil {
			return nil, fmt.Errorf("bad current config: %w", err)
		}
		// Assign input items to daeConfig.Global.
		inputGlobal.Assign(&c.Global)
		// Marshal back to string.
		marshaller := daeConfig.Marshaller{IndentSpace: 2}
		if err = marshaller.MarshalSection("global", reflect.ValueOf(c.Global), 0); err != nil {
			return nil, err
		}
		// This column should be updated.
		m.Global = string(marshaller.Bytes())
		updates["global"] = m.Global
	}
	if dns != nil {
		m.Dns = *dns
		updates["dns"] = m.Dns
	}
	if routing != nil {
		m.Routing = *routing
		updates["routing"] = m.Routing
	}
	// Check grammar.
	c, err := m.ToDaeConfig()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = tx.Model(&db.Config{ID: id}).Updates(updates).Error; err != nil {
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
	tx := db.BeginTx(ctx)
	m := db.Config{ID: id}
	q := tx.Clauses(clause.Returning{Columns: []clause.Column{{Name: "selected"}}}).
		Delete(&m)
	if q.Error != nil {
		tx.Rollback()
		return 0, q.Error
	}
	// Check if the config to delete is selected.
	if q.RowsAffected > 0 && m.Selected {
		// Check if dae is running.
		var sys db.System
		if err = tx.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
		if sys.Running {
			// Stop running with dry-run.
			if _, err = Run(tx, true); err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}
	tx.Commit()
	return int32(q.RowsAffected), nil
}

func Select(ctx context.Context, _id graphql.ID) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	// Unset all selected.
	q := tx.Model(&db.Config{}).Where("selected = ?", true).Update("selected", false)
	if err = q.Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	isReplace := q.RowsAffected > 0
	// Set selected.
	q = tx.Model(&db.Config{ID: id}).Update("selected", true)
	if err = q.Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if q.RowsAffected == 0 {
		tx.Rollback()
		return 0, fmt.Errorf("no such config")
	}
	if isReplace {
		// Check if dae is running.
		var sys db.System
		if err = tx.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
		if sys.Running {
			// Run with new config.
			if _, err = Run(tx, false); err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}
	tx.Commit()
	return 1, nil
}

func Run(tx *gorm.DB, dry bool) (n int32, err error) {
	// Dry run.
	if dry {
		ch := make(chan bool)
		dae.ChReloadConfigs <- &dae.ReloadMessage{
			Config:   dae.EmptyConfig,
			Callback: ch,
		}
		suc := <-ch
		if !suc {
			return 0, fmt.Errorf("failed to dryrun: unexpected failure; see more in log and report bugs")
		}

		// Running -> false
		var sys db.System
		if err = tx.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
			return 0, err
		}
		if err = tx.Model(&sys).Update("running", false).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
		return 1, nil
	}

	// Run selected config.
	var m db.Config
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
			switch name {
			case "direct", "block", "must_direct":
				// Preset groups.
			default:
				notFound = append(notFound, name)
			}
		}
		if len(notFound) > 0 {
			return 0, fmt.Errorf("groups not defined but referenced by routing: %v", strings.Join(notFound, ", "))
		}
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
		Find(&separateNodes, "subscription_id not in ?", subIds); err != nil {
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
	// Fill in group section.
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
	// Fill in node section.
	for name, node := range nameToNodes {
		c.Node = append(c.Node, daeConfig.KeyableString(fmt.Sprintf("%v:%v", name, node.Link)))
	}

	/// Reload with current config.
	chReloadCallback := make(chan bool)
	dae.ChReloadConfigs <- &dae.ReloadMessage{
		Config:   c,
		Callback: chReloadCallback,
	}
	sucReload := <-chReloadCallback
	if !sucReload {
		return 0, fmt.Errorf("failed to load new config; see more in log")
	}

	// Running -> true
	var sys db.System
	if err = tx.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
		return 0, err
	}
	if err = tx.Model(&sys).Update("running", true).Error; err != nil {
		return 0, err
	}

	return 1, nil
}
