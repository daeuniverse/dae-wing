/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

import (
	"context"
	"database/sql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
)

const (
	filename = "wing.db"
)

var (
	db *gorm.DB
)

func InitDatabase(configDir string) (err error) {
	db, err = gorm.Open(sqlite.Open(filepath.Join(configDir, filename)), &gorm.Config{})
	if err != nil {
		return err
	}
	if err = db.AutoMigrate(
		&Config{},
		&Node{},
		&Subscription{},
		&Group{},
		&GroupPolicyParamModel{},
	); err != nil {
		return err
	}
	return nil
}

func DB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}
func BeginTx(ctx context.Context) *gorm.DB {
	return DB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
}
