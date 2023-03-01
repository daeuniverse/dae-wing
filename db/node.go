/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

import (
	"database/sql"
	"gorm.io/gorm"
)

type Node struct {
	gorm.Model
	Link     string `gorm:"not null"`
	Name     string `gorm:"not null"`
	Address  string `gorm:"not null"`
	Protocol string `gorm:"not null"`

	Remarks string `gorm:"not null"`

	// Foreign keys.
	// Nil SubscriptionID indicates nodes belonging to no subscription.
	SubscriptionID sql.NullInt64
	Subscription   *Subscription
}
