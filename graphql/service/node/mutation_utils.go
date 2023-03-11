/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	"github.com/graph-gophers/graphql-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DuplicatedError = fmt.Errorf("node already exists")

type ImportResult struct {
	Link  string
	Error *string
	Node  *Resolver
}

func importNode(d *gorm.DB, subscriptionId *uint, arg *internal.ImportArgument) (m *db.Node, err error) {
	if err = arg.ValidateTag(); err != nil {
		return nil, err
	}
	m, err = db.NewNodeModel(arg.Link, arg.Tag, subscriptionId)
	if err != nil {
		return nil, err
	}
	var count int64
	if err = d.Model(&db.Node{}).
		Where("link = ?", arg.Link).
		Where("subscription_id = ?", subscriptionId).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, DuplicatedError
	}
	if err = d.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// Import nodes. If abortError is false, err will always be nil.
func Import(d *gorm.DB, abortError bool, subscriptionId *uint, argument []*internal.ImportArgument) (rs []*ImportResult, err error) {
	for _, arg := range argument {
		var m *db.Node
		if m, err = importNode(d, subscriptionId, arg); err != nil {
			if abortError && !errors.Is(err, DuplicatedError) {
				return nil, err
			}
			info := err.Error()
			rs = append(rs, &ImportResult{
				Link:  arg.Link,
				Error: &info,
				Node:  nil,
			})
			continue
		}
		rs = append(rs, &ImportResult{
			Link:  arg.Link,
			Error: nil,
			Node: &Resolver{
				Node: m,
			},
		})
	}
	return rs, nil
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
		Joins("inner join group_nodes on groups.system_id = ? and groups.id = group_nodes.group_id and group_nodes.node_id in ?", sys.ID, ids).
		Update("groups.version", gorm.Expr("groups.version + 1")).Error; err != nil {
		return err
	}

	return nil
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
		Delete(&db.Node{})
	if q.Error != nil {
		return 0, q.Error
	}

	// Update modified if any nodes are referenced by running config.
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
	q := db.DB(ctx).Model(&db.Node{}).
		Where("id = ?", id).
		Update("tag", tag)
	if q.Error != nil {
		return 0, q.Error
	}
	return int32(q.RowsAffected), nil
}
