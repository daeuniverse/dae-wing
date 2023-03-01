/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package node

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae-wing/graphql/service"
	"github.com/v2rayA/dae-wing/model"
)

type Resolver struct {
	*model.NodeModel
}

func (r *Resolver) Model() *service.ModelResolver {
	return &service.ModelResolver{
		Model: &r.NodeModel.Model,
	}
}
func (r *Resolver) Link() string {
	return r.NodeModel.Link
}
func (r *Resolver) Name() string {
	return r.NodeModel.Name
}
func (r *Resolver) Address() string {
	return r.NodeModel.Address
}
func (r *Resolver) Protocol() string {
	return r.NodeModel.Protocol
}
func (r *Resolver) Remarks() string {
	return r.NodeModel.Remarks
}
func (r *Resolver) Status() string {
	return r.NodeModel.Status
}
func (r *Resolver) SubscriptionID() *graphql.ID {
	return common.EncodeNullableCursor(r.NodeModel.SubscriptionID)
}
