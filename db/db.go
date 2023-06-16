/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daeuniverse/dae-wing/pkg/sqlite"
	"gorm.io/gorm"
)

const (
	filename = "wing.db"
)

var (
	db *gorm.DB
)

func InitDatabase(configDir string) (err error) {
	path := filepath.Join(configDir, filename)
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", err, path)
	}
	if err = db.AutoMigrate(
		&User{},
		&Config{},
		&Dns{},
		&Routing{},
		&Node{},
		&Subscription{},
		&Group{},
		&GroupPolicyParam{},
		&System{},
	); err != nil {
		return err
	}
	if fi, err := os.Stat(path); err != nil {
		return err
	} else if fi.Mode()&0037 > 0 {
		// Too open, chmod it to 0640.
		if err = os.Chmod(path, 0640); err != nil {
			return err
		}
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
func BeginReadOnlyTx(ctx context.Context) *gorm.DB {
	return DB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  true,
	})
}
