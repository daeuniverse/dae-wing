/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package subscription

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/service/node"
)

type Resolver struct {
	*db.Subscription
}

func (r *Resolver) ID() graphql.ID {
	return common.EncodeCursor(r.Subscription.ID)
}
func (r *Resolver) UpdatedAt() graphql.Time {
	return graphql.Time{
		Time: r.Subscription.UpdatedAt,
	}
}
func (r *Resolver) Tag() *string {
	return r.Subscription.Tag
}
func (r *Resolver) Link() string {
	return r.Subscription.Link
}
func (r *Resolver) Status() string {
	return r.Subscription.Status
}
func (r *Resolver) Info() string {
	return r.Subscription.Info
}
func (r *Resolver) Nodes(args *struct {
	First *int32
	After *graphql.ID
}) (*node.ConnectionResolver, error) {
	id := common.EncodeCursor(r.Subscription.ID)
	return node.NewConnectionResolver(nil, &id, args.First, args.After)
}
