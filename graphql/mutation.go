/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"github.com/graph-gophers/graphql-go"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/config"
	"github.com/daeuniverse/dae-wing/graphql/config/global"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	"github.com/daeuniverse/dae-wing/graphql/service/group"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/daeuniverse/dae-wing/graphql/service/subscription"
	"github.com/v2rayA/dae/pkg/config_parser"
)

type MutationResolver struct{}

func (r *MutationResolver) CreateConfig(args *struct {
	Global  *global.Input
	Dns     *string
	Routing *string
}) (c *config.Resolver, err error) {
	var strDns, strRouting string
	if args.Dns != nil {
		strDns = *args.Dns
	}
	if args.Routing != nil {
		strRouting = *args.Routing
	}
	return config.Create(context.TODO(), args.Global, strDns, strRouting)
}

func (r *MutationResolver) UpdateConfig(args *struct {
	ID      graphql.ID
	Global  *global.Input
	Dns     *string
	Routing *string
}) (*config.Resolver, error) {
	return config.Update(context.TODO(), args.ID, args.Global, args.Dns, args.Routing)
}

func (r *MutationResolver) RemoveConfig(args *struct {
	ID graphql.ID
}) (int32, error) {
	return config.Remove(context.TODO(), args.ID)
}

func (r *MutationResolver) SelectConfig(args *struct {
	ID graphql.ID
}) (int32, error) {
	return config.Select(context.TODO(), args.ID)
}

func (r *MutationResolver) Run(args *struct {
	Dry bool
}) (int32, error) {
	tx := db.BeginTx(context.TODO())
	ret, err := config.Run(tx, args.Dry)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()
	return ret, nil
}

func (r *MutationResolver) ImportNodes(args *struct {
	RollbackError bool
	Args          []*internal.ImportArgument
}) ([]*node.ImportResult, error) {
	tx := db.BeginTx(context.TODO())
	result, err := node.Import(tx, args.RollbackError, nil, args.Args)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return result, nil
}

func (r *MutationResolver) RemoveNodes(args *struct {
	IDs []graphql.ID
}) (int32, error) {
	return node.Remove(context.TODO(), args.IDs)
}

func (r *MutationResolver) TagNode(args *struct {
	ID  graphql.ID
	Tag string
}) (int32, error) {
	return node.Tag(context.TODO(), args.ID, args.Tag)
}

func (r *MutationResolver) ImportSubscription(args *struct {
	RollbackError bool
	Arg           internal.ImportArgument
}) (*subscription.ImportResult, error) {
	tx := db.BeginTx(context.TODO())
	result, err := subscription.Import(tx, args.RollbackError, &args.Arg)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return result, nil
}

func (r *MutationResolver) UpdateSubscription(args *struct {
	ID graphql.ID
}) (*subscription.Resolver, error) {
	return subscription.Update(context.TODO(), args.ID)
}

func (r *MutationResolver) RemoveSubscriptions(args *struct {
	IDs []graphql.ID
}) (int32, error) {
	return subscription.Remove(context.TODO(), args.IDs)
}

func (r *MutationResolver) TagSubscription(args *struct {
	ID  graphql.ID
	Tag string
}) (int32, error) {
	return subscription.Tag(context.TODO(), args.ID, args.Tag)
}

func (r *MutationResolver) CreateGroup(args *struct {
	Name         string
	Policy       string
	PolicyParams *[]config_parser.Param
}) (*group.Resolver, error) {
	var policyParams []config_parser.Param
	if args.PolicyParams != nil {
		policyParams = *args.PolicyParams
	}
	return group.Create(context.TODO(), args.Name, args.Policy, policyParams)
}

func (r *MutationResolver) RemoveGroup(args *struct {
	ID graphql.ID
}) (int32, error) {
	return group.Remove(context.TODO(), args.ID)
}

func (r *MutationResolver) RenameGroup(args *struct {
	ID   graphql.ID
	Name string
}) (int32, error) {
	return group.Rename(context.TODO(), args.ID, args.Name)
}

func (r *MutationResolver) GroupAddSubscriptions(args *struct {
	ID              graphql.ID
	SubscriptionIDs []graphql.ID
}) (int32, error) {
	return group.AddSubscriptions(context.TODO(), args.ID, args.SubscriptionIDs)
}

func (r *MutationResolver) GroupDelSubscriptions(args *struct {
	ID              graphql.ID
	SubscriptionIDs []graphql.ID
}) (int32, error) {
	return group.DelSubscriptions(context.TODO(), args.ID, args.SubscriptionIDs)
}

func (r *MutationResolver) GroupAddNodes(args *struct {
	ID      graphql.ID
	NodeIDs []graphql.ID
}) (int32, error) {
	return group.AddNodes(context.TODO(), args.ID, args.NodeIDs)
}

func (r *MutationResolver) GroupDelNodes(args *struct {
	ID      graphql.ID
	NodeIDs []graphql.ID
}) (int32, error) {
	return group.DelNodes(context.TODO(), args.ID, args.NodeIDs)
}
