/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package model

import (
	"context"
	"database/sql"
	"github.com/v2rayA/dae-wing/db"
	"gorm.io/gorm"
)

type NodeModel struct {
	gorm.Model
	Link     string `gorm:"not null"`
	Name     string `gorm:"not null"`
	Address  string `gorm:"not null"`
	Protocol string `gorm:"not null"`

	Remarks string `gorm:"not null"`
	Status  string `gorm:"not null"` // Error "unsupported" or something others.

	SubscriptionID sql.NullInt64
	Subscription   SubscriptionModel
}
type node struct{}

var Node node

func (node) Create(ctx context.Context, model *NodeModel) error {
	return db.DB(ctx).Create(model).Error
}
