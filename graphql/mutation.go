/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"database/sql"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/internal"
	"github.com/v2rayA/dae-wing/graphql/service/node"
	"github.com/v2rayA/dae-wing/graphql/service/subscription"
)

type MutationResolver struct{}

func (r *MutationResolver) ImportNodes(args *struct {
	RollbackError bool
	Args          []*internal.ImportArgument
}) ([]*node.ImportResult, error) {
	tx := db.DB(context.TODO()).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	result, err := node.Import(tx, args.RollbackError, nil, args.Args)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return result, nil
}

func (r *MutationResolver) ImportSubscription(args *struct {
	RollbackError bool
	Arg           internal.ImportArgument
}) ([]*node.ImportResult, error) {
	return subscription.Import(context.TODO(), args.RollbackError, &args.Arg)
}
