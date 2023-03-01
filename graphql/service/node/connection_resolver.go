/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package node

import (
	"context"
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/service"
	"gorm.io/gorm"
)

type ConnectionResolver struct {
	baseQuery func() *gorm.DB

	models []db.Node
}

func NewConnectionResolver(id *graphql.ID, subscriptionId *graphql.ID, first *int32, after *graphql.ID) (r *ConnectionResolver, err error) {
	var uintId, uintSubscriptionId uint
	if id != nil {
		uintId, err = common.DecodeCursor(*id)
		if err != nil {
			return nil, err
		}
	}
	if subscriptionId != nil {
		uintSubscriptionId, err = common.DecodeCursor(*subscriptionId)
		if err != nil {
			return nil, err
		}
	}
	baseQuery := func() *gorm.DB {
		q := db.DB(context.TODO()).Model(&db.Node{})
		if id != nil {
			q = q.Where("id = ?", uintId)
		}
		if subscriptionId != nil {
			q = q.Where("subscription_id = ?", uintSubscriptionId)
		}
		return q
	}

	q := baseQuery()
	if after != nil && first != nil {
		q = q.Where("id > ?", after).Limit(int(*first))
	}
	var models []db.Node
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	return &ConnectionResolver{
		baseQuery: baseQuery,
		models:    models,
	}, nil
}

func (r *ConnectionResolver) TotalCount() (int32, error) {
	var count int64
	if err := r.baseQuery().Count(&count).Error; err != nil {
		return 0, err
	}
	return int32(count), nil
}
func (r *ConnectionResolver) Edges() (rs []*Resolver, err error) {
	for _, _m := range r.models {
		m := _m
		rs = append(rs, &Resolver{
			Node: &m,
		})
	}
	return rs, nil
}
func (r *ConnectionResolver) PageInfo() (pr *service.PageInfoResolver, err error) {
	if len(r.models) == 0 {
		return &service.PageInfoResolver{
			FStartCursor: nil,
			FEndCursor:   nil,
			FHasNextPage: false,
		}, nil
	}
	start := common.EncodeCursor(r.models[0].ID)
	end := common.EncodeCursor(r.models[len(r.models)-1].ID)
	// Get the last ID.
	var lastNode db.Node
	if err := r.baseQuery().Select("id").Order("id DESC").First(&lastNode).Error; err != nil {
		return nil, err
	}
	return &service.PageInfoResolver{
		FStartCursor: &start,
		FEndCursor:   &end,
		FHasNextPage: r.models[len(r.models)-1].ID < lastNode.ID,
	}, nil
}
