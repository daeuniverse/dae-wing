/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package routing

import (
	"context"
	"fmt"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/service/config"
	"github.com/graph-gophers/graphql-go"
	daeConfig "github.com/v2rayA/dae/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

func Create(ctx context.Context, name string, routing string) (*Resolver, error) {
	routing = "routing {\n" + routing + "\n}"
	m := db.Routing{
		ID:       0,
		Name:     name,
		Routing:  routing,
		Selected: false,
	}
	// Check grammar and to dae config.
	c, err := dae.ParseConfig(nil, nil, &m.Routing)
	if err != nil {
		return nil, err
	}
	if err = db.DB(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return &Resolver{
		DaeRouting: &c.Routing,
		Model:      &m,
	}, nil
}

func Update(ctx context.Context, _id graphql.ID, routing string) (*Resolver, error) {
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
	var m db.Routing
	if err = tx.Model(&db.Routing{}).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	// Prepare to partially update.
	// Convert routing string in database to daeConfig.Routing.
	c, err := dae.ParseConfig(nil, nil, &m.Routing)
	if err != nil {
		return nil, fmt.Errorf("bad current routing: %w", err)
	}
	m.Routing = "routing {\n" + routing + "\n}"
	// Marshal back to string.
	marshaller := daeConfig.Marshaller{IndentSpace: 2}
	if err = marshaller.MarshalSection("routing", reflect.ValueOf(c.Routing), 0); err != nil {
		return nil, err
	}
	// Update.
	if err = tx.Model(&db.Routing{ID: id}).Updates(map[string]interface{}{
		"routing": string(marshaller.Bytes()),
		"version": gorm.Expr("version + 1"),
	}).Error; err != nil {
		return nil, err
	}
	return &Resolver{
		DaeRouting: &c.Routing,
		Model:      &m,
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
	m := db.Routing{ID: id}
	q := tx.Clauses(clause.Returning{Columns: []clause.Column{{Name: "selected"}}}).
		Select(clause.Associations).
		Delete(&m)
	if q.Error != nil {
		return 0, q.Error
	}
	// Check if the config to delete is selected.
	if q.RowsAffected > 0 && m.Selected {
		// Check if dae is running.
		var sys db.System
		if err = tx.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
			return 0, err
		}
		if sys.Running {
			// Stop running with dry-run.
			if _, err = config.Run(tx, true); err != nil {
				return 0, err
			}
		}
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
	q := tx.Model(&db.Routing{}).Where("selected = ?", true).Update("selected", false)
	if err = q.Error; err != nil {
		return 0, err
	}
	isReplace := q.RowsAffected > 0
	// Set selected.
	q = tx.Model(&db.Routing{ID: id}).Update("selected", true)
	if err = q.Error; err != nil {
		return 0, err
	}
	if q.RowsAffected == 0 {
		return 0, fmt.Errorf("no such config")
	}
	if isReplace {
		// Check if dae is running.
		var sys db.System
		if err = tx.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
			return 0, err
		}
		if sys.Running {
			// Run with new config.
			if _, err = config.Run(tx, false); err != nil {
				return 0, err
			}
		}
	}
	return 1, nil
}

func Rename(ctx context.Context, _id graphql.ID, name string) (n int32, err error) {
	id, err := common.DecodeCursor(_id)
	if err != nil {
		return 0, err
	}
	q := db.DB(ctx).Model(&db.Routing{ID: id}).
		Update("name", name)
	if q.Error != nil {
		return 0, q.Error
	}
	return int32(q.RowsAffected), nil
}
