/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package group

import (
	"context"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/config"
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae/pkg/config_parser"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Create(ctx context.Context, name string, policy string, policyParams []config_parser.Param) (r *Resolver, err error) {
	if err = common.ValidateId(name); err != nil {
		return nil, err
	}
	params := make([]db.GroupPolicyParam, len(policyParams))
	for i := range params {
		params[i].Unmarshal(&policyParams[i])
	}
	m := db.Group{
		ID:           0,
		Name:         name,
		Policy:       policy,
		PolicyParams: params,
	}
	if err = db.DB(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return &Resolver{
		Group: &m,
	}, nil
}

func referenceGroups(d *gorm.DB) (groups []string, err error) {
	var sys db.System
	if err = d.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
		return nil, err
	}

	if !sys.Running {
		return nil, nil
	}

	var conf db.Config
	if err = d.Model(&db.Config{}).Where("selected = ?", true).First(&conf).Error; err != nil {
		return nil, err
	}
	daeConf, err := conf.ToDaeConfig()
	if err != nil {
		return nil, err
	}
	groups = config.NecessaryOutbounds(&daeConf.Routing)
	return groups, nil
}

func inReferenceGroups(d *gorm.DB, groupName string) (bool, error) {
	groups, err := referenceGroups(d)
	if err != nil {
		return false, err
	}
	for _, g := range groups {
		if g == groupName {
			return true, nil
		}
	}
	return false, nil
}

func setModified(d *gorm.DB) (err error) {
	var sys db.System
	// Get ID.
	if err = d.Model(&sys).Select("id").FirstOrCreate(&sys).Error; err != nil {
		return err
	}
	return d.Model(&sys).Update("modified", true).Error
}

func autoUpdateModifiedById(d *gorm.DB, id uint) (err error) {
	// Set modified = true if the group is referenced by selected config.
	var g db.Group
	if err = d.Select("name").Model(&db.Group{ID: id}).First(&g).Error; err != nil {
		return err
	}
	return autoUpdateModifiedByName(d, g.Name)
}

func autoUpdateModifiedByName(d *gorm.DB, name string) (err error) {
	// Set modified = true if the group is referenced by selected config.
	isReferenced, err := inReferenceGroups(d, name)
	if err != nil {
		return err
	}
	if isReferenced {
		if err = setModified(d); err != nil {
			return err
		}
	}
	return nil
}

func Rename(ctx context.Context, _id graphql.ID, name string) (n int32, err error) {
	if err = common.ValidateId(name); err != nil {
		return 0, err
	}
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	g := db.Group{ID: id}
	if err = tx.Model(&g).First(&g).Error; err != nil {
		return 0, err
	}
	q := tx.Model(&g).Update("name", name)
	if q.Error != nil {
		return 0, err
	}
	// Set modified = true if the group is changed and referenced by selected config.
	if q.Statement.Changed() {
		if err = autoUpdateModifiedByName(tx, g.Name); err != nil {
			return 0, err
		}
	}

	return int32(q.RowsAffected), nil
}

func Remove(ctx context.Context, _id graphql.ID) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	g := db.Group{ID: id}
	q := tx.Select(clause.Associations).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "name"}}}).
		Delete(&g)
	if q.Error != nil {
		return 0, q.Error
	}
	if err = tx.Model(&db.Group{ID: id}).Association("Node").Clear(); err != nil {
		return 0, err
	}
	if err = tx.Model(&db.Group{ID: id}).Association("Subscription").Clear(); err != nil {
		return 0, err
	}
	if err = tx.Where("group_id = ?", id).Delete(&db.GroupPolicyParam{}).Error; err != nil {
		return 0, err
	}
	// Set modified = true if the group is referenced by selected config.
	if err = autoUpdateModifiedByName(tx, g.Name); err != nil {
		return 0, err
	}
	return int32(q.RowsAffected), nil
}

func AddSubscriptions(ctx context.Context, _id graphql.ID, _subscriptionIds []graphql.ID) (int32, error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	subscriptionIds, err := common.DecodeCursorBatch(_subscriptionIds)
	if err != nil {
		return 0, err
	}
	var subs []db.Subscription
	for _, id := range subscriptionIds {
		subs = append(subs, db.Subscription{ID: id})
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	if err = tx.Model(&db.Group{ID: id}).
		Association("Subscription").
		Append(subs); err != nil {
		return 0, err
	}

	// Set modified = true if the group is referenced by selected config.
	if err = autoUpdateModifiedById(tx, id); err != nil {
		return 0, err
	}
	return int32(len(subscriptionIds)), nil
}

func DelSubscriptions(ctx context.Context, _id graphql.ID, _subscriptionIds []graphql.ID) (int32, error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	subscriptionIds, err := common.DecodeCursorBatch(_subscriptionIds)
	if err != nil {
		return 0, err
	}
	var subs []db.Subscription
	for _, id := range subscriptionIds {
		subs = append(subs, db.Subscription{ID: id})
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	if err = tx.Model(&db.Group{ID: id}).
		Association("Subscription").
		Delete(subs); err != nil {
		return 0, err
	}

	// Set modified = true if the group is referenced by selected config.
	if err = autoUpdateModifiedById(tx, id); err != nil {
		return 0, err
	}
	return int32(len(subscriptionIds)), nil
}

func AddNodes(ctx context.Context, _id graphql.ID, _nodeIds []graphql.ID) (int32, error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	nodeIds, err := common.DecodeCursorBatch(_nodeIds)
	if err != nil {
		return 0, err
	}
	var nodes []db.Node
	for _, id := range nodeIds {
		nodes = append(nodes, db.Node{ID: id})
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	if err = tx.Model(&db.Group{ID: id}).
		Association("Node").
		Append(nodes); err != nil {
		return 0, err
	}

	// Set modified = true if the group is referenced by selected config.
	if err = autoUpdateModifiedById(tx, id); err != nil {
		return 0, err
	}
	return int32(len(_nodeIds)), nil
}

func DelNodes(ctx context.Context, _id graphql.ID, _nodeIds []graphql.ID) (int32, error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	nodeIds, err := common.DecodeCursorBatch(_nodeIds)
	if err != nil {
		return 0, err
	}
	var nodes []db.Node
	for _, id := range nodeIds {
		nodes = append(nodes, db.Node{ID: id})
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	if err = tx.Model(&db.Group{ID: id}).
		Association("Node").
		Delete(nodes); err != nil {
		return 0, err
	}

	// Set modified = true if the group is referenced by selected config.
	if err = autoUpdateModifiedById(tx, id); err != nil {
		return 0, err
	}
	return int32(len(_nodeIds)), nil
}
