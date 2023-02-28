/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package subscription

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/graphql/service"
	"github.com/v2rayA/dae-wing/graphql/service/node"
	"github.com/v2rayA/dae-wing/model"
)

type Resolver struct {
	*model.SubscriptionModel
}

func (r *Resolver) Model() *service.ModelResolver {
	return &service.ModelResolver{
		Model: &r.SubscriptionModel.Model,
	}
}
func (r *Resolver) Remarks() string {
	return r.SubscriptionModel.Remarks
}
func (r *Resolver) Link() string {
	return r.SubscriptionModel.Link
}
func (r *Resolver) Status() string {
	return r.SubscriptionModel.Status
}
func (r *Resolver) Info() string {
	return r.SubscriptionModel.Info
}
func (r *Resolver) Nodes(args struct {
	First *int32
	After *graphql.ID
}) (*node.ConnectionResolver, error) {
	return node.NewConnectionResolver(r.SubscriptionModel.ID, args.First, args.After)
}
