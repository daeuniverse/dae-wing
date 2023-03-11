/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

import (
	"github.com/v2rayA/dae/pkg/config_parser"
)

type Group struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	Name         string `gorm:"not null;unique;index"`
	Policy       string `gorm:"not null"`
	PolicyParams []GroupPolicyParam
	Node         []Node         `gorm:"many2many:group_nodes;"`
	Subscription []Subscription `gorm:"many2many:group_subscriptions;"`

	Version  uint `gorm:"not null;default:0"`
	SystemID *uint
}

type GroupPolicyParam struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Key   string `gorm:"not null"`
	Value string `gorm:"not null"`

	// Foreign keys.
	GroupID uint
	Group   Group
}

func (m *GroupPolicyParam) Marshal() *config_parser.Param {
	return &config_parser.Param{
		Key: m.Key,
		Val: m.Value,
	}
}

func (m *GroupPolicyParam) Unmarshal(param *config_parser.Param) {
	m.Key = param.Key
	m.Value = param.Val
}
