/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package group

import (
	"context"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
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

func autoUpdateVersionById(d *gorm.DB, id uint) (err error) {
	return d.Model(db.Group{ID: id}).
		Update("version", gorm.Expr("version + 1")).Error
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
		if err = autoUpdateVersionById(tx, g.ID); err != nil {
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
	if err = autoUpdateVersionById(tx, g.ID); err != nil {
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

	if err = autoUpdateVersionById(tx, id); err != nil {
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

	if err = autoUpdateVersionById(tx, id); err != nil {
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

	if err = autoUpdateVersionById(tx, id); err != nil {
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

	if err = autoUpdateVersionById(tx, id); err != nil {
		return 0, err
	}
	return int32(len(_nodeIds)), nil
}

func SetPolicy(ctx context.Context, _id graphql.ID, policy string, policyParams []config_parser.Param) (n int32, err error) {
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
	q := tx.Model(&db.Group{ID: id}).Update("policy", policy)
	if err = q.Error; err != nil {
		return 0, err
	}
	if q.RowsAffected == 0 {
		return 0, nil
	}
	params := make([]db.GroupPolicyParam, len(policyParams))
	for i := range params {
		params[i].Unmarshal(&policyParams[i])
	}
	if err = tx.Model(&db.Group{ID: id}).Association("PolicyParams").Replace(params); err != nil {
		return 0, err
	}

	if err = autoUpdateVersionById(tx, id); err != nil {
		return 0, err
	}

	return int32(q.RowsAffected), nil
}
