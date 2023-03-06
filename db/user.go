/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

import (
	"context"
)

type User struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
}
type user struct{}

var UserInstance user

func (user) Create(ctx context.Context, model *User) error {
	return DB(ctx).Create(model).Error
}

func (user) Exists(ctx context.Context, model *User) (bool, error) {
	var count int64
	if err := DB(ctx).Model(model).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
