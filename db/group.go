/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

import (
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	Name           string `gorm:"not null;unique;index"`
	Policy         string `gorm:"not null"`
	StrategyParams []GroupStrategyParamModel
	Node           []Node         `gorm:"many2many:group_nodes;"`
	Subscription   []Subscription `gorm:"many2many:group_subscriptions;"`
}

type GroupStrategyParamModel struct {
	gorm.Model
	Key   string `gorm:"not null"`
	Value string `gorm:"not null"`

	GroupID uint `gorm:"not null"`
	Group   Group
}
