/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package node

import (
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/internal"
	"gorm.io/gorm"
)

type ImportResult struct {
	Link  string
	Error *string
	Node  *Resolver
}

func importNode(d *gorm.DB, subscriptionId *uint, arg *internal.ImportArgument) (m *db.Node, err error) {
	m, err = db.NewNodeModel(arg.Link, arg.Remarks, subscriptionId)
	if err != nil {
		return nil, err
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
			if abortError {
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
