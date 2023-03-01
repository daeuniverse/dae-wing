/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package subscription

import (
	"context"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/internal"
	"github.com/v2rayA/dae-wing/graphql/service/node"
	"github.com/v2rayA/dae/common/subscription"
	"gorm.io/gorm/clause"
	"io"
	"net/http"
	"time"
)

type ImportResult struct {
	Link         string
	Error        *string
	Subscription *Resolver
}

func Import(ctx context.Context, rollbackError bool, argument *internal.ImportArgument) (rs *ImportResult, err error) {
	if err = argument.ValidateTag(); err != nil {
		return nil, err
	}
	/// Resolve subscription to node links.
	// Fetch subscription link.
	var (
		b    []byte
		resp *http.Response
	)
	resp, err = http.Get(argument.Link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Resolve node links.
	noLogger := logrus.New()
	noLogger.SetOutput(io.Discard)
	links, err := subscription.ResolveSubscriptionAsSIP008(noLogger, b)
	if err != nil {
		links = subscription.ResolveSubscriptionAsBase64(noLogger, b)
	}
	/// Create a subscription model.
	tx := db.BeginTx(ctx)
	m := db.Subscription{
		ID:        0,
		UpdatedAt: time.Now(),
		Tag:       argument.Tag,
		Link:      argument.Link,
		Status:    "",
		Info:      "", // not supported yet
		Node:      nil,
	}
	if err = tx.Create(&m).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	/// Import nodes.
	// Links to import arguments.
	var args []*internal.ImportArgument
	for _, link := range links {
		args = append(args, &internal.ImportArgument{
			Link: link,
			Tag:  nil,
		})
	}
	// Import nodes.
	_, err = node.Import(tx, rollbackError, &m.ID, args)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &ImportResult{
		Link:  argument.Link,
		Error: nil,
		Subscription: &Resolver{
			Subscription: &m,
		},
	}, nil
}

func Remove(ctx context.Context, _ids []graphql.ID) (n int32, err error) {
	ids, err := common.DecodeCursorBatch(_ids)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	if err = tx.Where("id in ?", ids).
		Select(clause.Associations).
		Delete(&db.Subscription{}).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Where("subscription_id in ?", ids).Delete(&db.Node{}).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()
	return int32(len(ids)), nil
}

func Tag(ctx context.Context, _id graphql.ID, tag string) (n int32, err error) {
	if err = common.ValidateTag(tag); err != nil {
		return 0, err
	}
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	if err = db.DB(ctx).Model(&db.Subscription{}).
		Where("id = ?", id).
		Update("tag", tag).Error; err != nil {
		return 0, err
	}
	return 1, nil
}
