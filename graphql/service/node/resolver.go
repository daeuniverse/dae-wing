/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package node

import (
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/graph-gophers/graphql-go"
)

type Resolver struct {
	*db.Node
}

func (r *Resolver) ID() graphql.ID {
	return common.EncodeCursor(r.Node.ID)
}
func (r *Resolver) Link() string {
	return r.Node.Link
}
func (r *Resolver) Name() string {
	return r.Node.Name
}
func (r *Resolver) Address() string {
	return r.Node.Address
}
func (r *Resolver) Protocol() string {
	return r.Node.Protocol
}
func (r *Resolver) Tag() *string {
	return r.Node.Tag
}
func (r *Resolver) SubscriptionID() *graphql.ID {
	return common.EncodeNullableCursor(r.Node.SubscriptionID)
}
