/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package service

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"gorm.io/gorm"
)

type ModelResolver struct {
	*gorm.Model
}

func (r *ModelResolver) ID() (id graphql.ID) {
	return common.EncodeCursor(r.Model.ID)
}

func (r *ModelResolver) CreatedAt() graphql.Time {
	return graphql.Time{
		Time: r.Model.CreatedAt,
	}
}

func (r *ModelResolver) UpdatedAt() graphql.Time {
	return graphql.Time{
		Time: r.Model.UpdatedAt,
	}
}

func (r *ModelResolver) DeletedAt() (t *graphql.Time) {
	if !r.Model.DeletedAt.Valid {
		return nil
	}
	return &graphql.Time{
		Time: r.Model.DeletedAt.Time,
	}

}
