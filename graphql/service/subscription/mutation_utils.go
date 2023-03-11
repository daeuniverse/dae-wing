/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package subscription

import (
	"context"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
	"github.com/v2rayA/dae/common/subscription"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"net/http"
	"time"
)

type ImportResult struct {
	Link             string
	NodeImportResult []*node.ImportResult
	Sub              *Resolver
}

func fetchLinks(subscriptionLink string) (links []string, err error) {
	/// Resolve subscription to node links.
	// Fetch subscription link.
	var (
		b    []byte
		resp *http.Response
	)
	resp, err = http.Get(subscriptionLink)
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
	links, err = subscription.ResolveSubscriptionAsSIP008(noLogger, b)
	if err != nil {
		links = subscription.ResolveSubscriptionAsBase64(noLogger, b)
	}
	return links, nil
}

func Import(c *gorm.DB, rollbackError bool, argument *internal.ImportArgument) (r *ImportResult, err error) {
	if err = argument.ValidateTag(); err != nil {
		return nil, err
	}
	links, err := fetchLinks(argument.Link)
	if err != nil {
		return nil, err
	}
	/// Create a subscription model.
	m := db.Subscription{
		ID:        0,
		UpdatedAt: time.Now(),
		Tag:       argument.Tag,
		Link:      argument.Link,
		Status:    "",
		Info:      "", // not supported yet
		Node:      nil,
	}
	if err = c.Create(&m).Error; err != nil {
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
	result, err := node.Import(c, rollbackError, &m.ID, args)
	if err != nil {
		return nil, err
	}
	return &ImportResult{
		Link:             argument.Link,
		NodeImportResult: result,
		Sub: &Resolver{
			Subscription: &m,
		},
	}, nil
}

func autoUpdateVersionByIds(d *gorm.DB, ids []uint) (err error) {
	var sys db.System
	if err = d.Model(&db.System{}).
		FirstOrCreate(&sys).Error; err != nil {
		return err
	}
	if !sys.Running {
		return nil
	}

	if err = d.Model(&db.Group{}).
		Joins("inner join group_subscriptions on groups.system_id = ? and groups.id = group_subscriptions.group_id and group_subscriptions.subscription_id in ?", sys.ID, ids).
		Update("groups.version", gorm.Expr("groups.version + 1")).Error; err != nil {
		return err
	}

	return nil
}

func Update(ctx context.Context, _id graphql.ID) (r *Resolver, err error) {
	subId, err := common.DecodeCursor(_id)
	if err != nil {
		return nil, err
	}
	// Fetch node links.
	var m db.Subscription
	if err = db.DB(ctx).Where(&db.Subscription{ID: subId}).First(&m).Error; err != nil {
		return nil, err
	}
	links, err := fetchLinks(m.Link)
	if err != nil {
		return nil, err
	}

	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	// Remove those subscription_id of which satisfied and are not independently in any groups.
	subQuery := tx.Model(&db.Node{}).
		Where("subscription_id = ?", subId).
		Select("nodes.id as id").
		Joins("inner join group_nodes on group_nodes.node_id = nodes.id")

	if err = tx.Where("subscription_id = ?", subId).
		Where("id not in (?)", subQuery).
		Select(clause.Associations).
		Delete(&db.Node{}).Error; err != nil {
		return nil, err
	}
	// Import node links.
	var args []*internal.ImportArgument
	for _, link := range links {
		args = append(args, &internal.ImportArgument{Link: link})
	}
	if _, err = node.Import(tx, false, &subId, args); err != nil {
		return nil, err
	}
	// Update updated_at and return the latest version.
	if err = tx.Model(&m).
		Clauses(clause.Returning{}).
		Where(&db.Subscription{ID: subId}).
		Update("updated_at", time.Now()).Error; err != nil {
		return nil, err
	}

	// Update modified if subscription is referenced by running config.
	if err = autoUpdateVersionByIds(tx, []uint{subId}); err != nil {
		return nil, err
	}
	return &Resolver{Subscription: &m}, nil
}

func Remove(ctx context.Context, _ids []graphql.ID) (n int32, err error) {
	ids, err := common.DecodeCursorBatch(_ids)
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
	q := tx.Where("id in ?", ids).
		Select(clause.Associations).
		Delete(&db.Subscription{})
	if q.Error != nil {
		return 0, q.Error
	}
	if err = tx.Where("subscription_id in ?", ids).Delete(&db.Node{}).Error; err != nil {
		return 0, err
	}

	// Update modified if any subscriptions are referenced by running config.
	if err = autoUpdateVersionByIds(tx, ids); err != nil {
		return 0, err
	}

	return int32(q.RowsAffected), nil
}

func Tag(ctx context.Context, _id graphql.ID, tag string) (n int32, err error) {
	if err = common.ValidateTag(tag); err != nil {
		return 0, err
	}
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	q := db.DB(ctx).Model(&db.Subscription{}).
		Where("id = ?", id).
		Update("tag", tag)
	if q.Error != nil {
		return 0, q.Error
	}
	return int32(q.RowsAffected), nil
}
