/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package group

import (
	"context"
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae/pkg/config_parser"
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

func Rename(ctx context.Context, _id graphql.ID, name string) (n int32, err error) {
	if err = common.ValidateId(name); err != nil {
		return 0, err
	}
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	q := db.DB(ctx).Model(&db.Group{}).
		Where("id = ?", id).
		Update("name", name)
	if q.Error != nil {
		return 0, err
	}
	return int32(q.RowsAffected), nil
}

func Remove(ctx context.Context, _id graphql.ID) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	q := tx.Select(clause.Associations).
		Delete(&db.Group{}, "id = ?", id)
	if q.Error != nil {
		tx.Rollback()
		return 0, q.Error
	}
	if err = tx.Model(&db.Group{ID: id}).Association("Node").Clear(); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Model(&db.Group{ID: id}).Association("Subscription").Clear(); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Where("group_id = ?", id).Delete(&db.GroupPolicyParam{}).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()
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
	if err = db.DB(ctx).
		Model(&db.Group{ID: id}).
		Association("Subscription").
		Append(subs); err != nil {
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
	if err = db.DB(ctx).
		Model(&db.Group{ID: id}).
		Association("Subscription").
		Delete(subs); err != nil {
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
	if err = db.DB(ctx).
		Model(&db.Group{ID: id}).
		Association("Node").
		Append(nodes); err != nil {
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
	if err = db.DB(ctx).
		Model(&db.Group{ID: id}).
		Association("Node").
		Delete(nodes); err != nil {
		return 0, err
	}
	return int32(len(_nodeIds)), nil
}
