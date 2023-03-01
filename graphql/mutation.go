/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"github.com/v2rayA/dae-wing/graphql/service/node"
)

type MutationResolver struct{}

func (r *MutationResolver) ImportNodes(args *struct {
	RollbackError bool
	Nodes         []*node.ImportArgument
}) ([]*node.ImportResult, error) {
	return node.ImportNodes(context.TODO(), args.RollbackError, args.Nodes)
}
