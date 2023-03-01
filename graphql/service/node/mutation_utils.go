/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package node

import (
	"context"
	"database/sql"
	"github.com/v2rayA/dae-wing/db"
	"gorm.io/gorm"
)

type ImportArgument struct {
	Link    string
	Remarks *string
}
type ImportResult struct {
	Error *string
	Node  *Resolver
}

func importNode(gormDb *gorm.DB, arg *ImportArgument) (m *db.Node, err error) {
	remarks := ""
	if arg.Remarks != nil {
		remarks = *arg.Remarks
	}
	m, err = db.NewNodeModel(arg.Link, remarks, sql.NullInt64{})
	if err != nil {
		return nil, err
	}
	if err = gormDb.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func ImportNodes(ctx context.Context, rollbackError bool, argument []*ImportArgument) (rs []*ImportResult, err error) {
	tx := db.DB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	for _, arg := range argument {
		var m *db.Node
		if m, err = importNode(tx, arg); err != nil {
			if rollbackError {
				tx.Rollback()
				return nil, err
			}
			info := err.Error()
			rs = append(rs, &ImportResult{
				Error: &info,
				Node:  nil,
			})
			continue
		}
		rs = append(rs, &ImportResult{
			Error: nil,
			Node: &Resolver{
				Node: m,
			},
		})
	}
	tx.Commit()
	return rs, nil
}
