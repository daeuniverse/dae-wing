/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package model

import (
	"context"
	"github.com/v2rayA/dae-wing/db"
	"gorm.io/gorm"
)

type UserModel struct {
	gorm.Model
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
}
type user struct{}

var User user

func (user) Create(ctx context.Context, model *UserModel) error {
	return db.DB(ctx).Create(model).Error
}

func (user) Exists(ctx context.Context, model *UserModel) (bool, error) {
	var count int64
	if err := db.DB(ctx).Model(model).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
