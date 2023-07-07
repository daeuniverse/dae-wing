/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dns

import (
	"context"
	"fmt"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/graph-gophers/graphql-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Create(ctx context.Context, name string, dns string) (*Resolver, error) {
	dns = "dns {\n" + dns + "\n}"
	m := db.Dns{
		ID:       0,
		Name:     name,
		Dns:      dns,
		Selected: false,
	}
	// Check grammar and to dae config.
	c, err := dae.ParseConfig(nil, &m.Dns, nil)
	if err != nil {
		return nil, err
	}
	if err = db.DB(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return &Resolver{
		DaeDns: &c.Dns,
		Model:  &m,
	}, nil
}

func Update(ctx context.Context, _id graphql.ID, dns string) (*Resolver, error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return nil, err
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	var m db.Dns
	if err = tx.Model(&db.Dns{}).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	// Prepare to partially update.
	m.Dns = "dns {\n" + dns + "\n}"
	// Parse it check the grammar.
	c, err := dae.ParseConfig(nil, &m.Dns, nil)
	if err != nil {
		return nil, fmt.Errorf("bad current dns: %w", err)
	}
	// Update.
	if err = tx.Model(&db.Dns{ID: id}).Updates(map[string]interface{}{
		"dns":     m.Dns,
		"version": gorm.Expr("version + 1"),
	}).Error; err != nil {
		return nil, err
	}
	return &Resolver{
		DaeDns: &c.Dns,
		Model:  &m,
	}, nil
}

func Remove(ctx context.Context, _id graphql.ID) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	m := db.Dns{ID: id}
	q := tx.Clauses(clause.Returning{Columns: []clause.Column{{Name: "selected"}}}).
		Select(clause.Associations).
		Delete(&m)
	if q.Error != nil {
		return 0, q.Error
	}
	return int32(q.RowsAffected), nil
}

func Select(ctx context.Context, _id graphql.ID) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	tx := db.BeginTx(ctx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	// Unset all selected.
	q := tx.Model(&db.Dns{}).Where("selected = ?", true).Update("selected", false)
	if err = q.Error; err != nil {
		return 0, err
	}
	// Set selected.
	q = tx.Model(&db.Dns{ID: id}).Update("selected", true)
	if err = q.Error; err != nil {
		return 0, err
	}
	if q.RowsAffected == 0 {
		return 0, fmt.Errorf("no such config")
	}
	return 1, nil
}

func Rename(ctx context.Context, _id graphql.ID, name string) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	q := db.DB(ctx).Model(&db.Dns{ID: id}).
		Update("name", name)
	if q.Error != nil {
		return 0, q.Error
	}
	return int32(q.RowsAffected), nil
}
