/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package subscription

import (
	"context"
	"database/sql"
	"github.com/sirupsen/logrus"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/internal"
	"github.com/v2rayA/dae-wing/graphql/service/node"
	"github.com/v2rayA/dae/common/subscription"
	"gorm.io/gorm"
	"io"
	"net/http"
)

func Import(ctx context.Context, rollbackError bool, argument *internal.ImportArgument) (rs []*node.ImportResult, err error) {
	/// Resolve subscription to node links.
	// Fetch subscription link.
	var (
		b    []byte
		resp *http.Response
	)
	resp, err = http.Get(argument.Link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Resolve node links.
	noLogger := logrus.New()
	noLogger.SetOutput(io.Discard)
	links, err := subscription.ResolveSubscriptionAsSIP008(noLogger, b)
	if err != nil {
		links = subscription.ResolveSubscriptionAsBase64(noLogger, b)
	}
	// Links to import arguments.
	var args []*internal.ImportArgument
	for _, link := range links {
		args = append(args, &internal.ImportArgument{
			Link:    link,
			Remarks: nil,
		})
	}
	// Create a subscription model.
	tx := db.DB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	m := db.Subscription{
		Model:   gorm.Model{},
		Remarks: argument.Remarks,
		Link:    argument.Link,
		Status:  "",
		Info:    "", // not supported yet
		Nodes:   nil,
	}
	if err = tx.Create(&m).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	// Import nodes.
	result, err := node.Import(tx, rollbackError, &m.ID, args)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return result, nil
}
