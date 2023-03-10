/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	"github.com/daeuniverse/dae-wing/graphql/service/config"
	"github.com/daeuniverse/dae-wing/graphql/service/config/global"
	"github.com/daeuniverse/dae-wing/graphql/service/dns"
	"github.com/daeuniverse/dae-wing/graphql/service/group"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/daeuniverse/dae-wing/graphql/service/routing"
	"github.com/daeuniverse/dae-wing/graphql/service/subscription"
	"github.com/graph-gophers/graphql-go"
	"github.com/v2rayA/dae/pkg/config_parser"
)

type MutationResolver struct{}

func (r *MutationResolver) CreateConfig(args *struct {
	Name   *string
	Global *global.Input
}) (c *config.Resolver, err error) {
	var strName string
	if args.Name != nil {
		strName = *args.Name
	}
	return config.Create(context.TODO(), strName, args.Global)
}

func (r *MutationResolver) UpdateConfig(args *struct {
	ID     graphql.ID
	Global global.Input
}) (*config.Resolver, error) {
	return config.Update(context.TODO(), args.ID, args.Global)
}

func (r *MutationResolver) RenameConfig(args *struct {
	ID   graphql.ID
	Name string
}) (int32, error) {
	return config.Rename(context.TODO(), args.ID, args.Name)
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

func (r *MutationResolver) CreateDns(args *struct {
	Name *string
	Dns  *string
}) (c *dns.Resolver, err error) {
	var strDns, strName string
	if args.Dns != nil {
		strDns = *args.Dns
	}
	if args.Name != nil {
		strName = *args.Name
	}
	return dns.Create(context.TODO(), strName, strDns)
}

func (r *MutationResolver) UpdateDns(args *struct {
	ID  graphql.ID
	Dns string
}) (*dns.Resolver, error) {
	return dns.Update(context.TODO(), args.ID, args.Dns)
}

func (r *MutationResolver) RenameDns(args *struct {
	ID   graphql.ID
	Name string
}) (int32, error) {
	return dns.Rename(context.TODO(), args.ID, args.Name)
}

func (r *MutationResolver) RemoveDns(args *struct {
	ID graphql.ID
}) (int32, error) {
	return dns.Remove(context.TODO(), args.ID)
}

func (r *MutationResolver) SelectDns(args *struct {
	ID graphql.ID
}) (int32, error) {
	return dns.Select(context.TODO(), args.ID)
}

func (r *MutationResolver) CreateRouting(args *struct {
	Name    *string
	Routing *string
}) (c *routing.Resolver, err error) {
	var strRouting, strName string
	if args.Routing != nil {
		strRouting = *args.Routing
	}
	if args.Name != nil {
		strName = *args.Name
	}
	return routing.Create(context.TODO(), strName, strRouting)
}

func (r *MutationResolver) UpdateRouting(args *struct {
	ID      graphql.ID
	Routing string
}) (*routing.Resolver, error) {
	return routing.Update(context.TODO(), args.ID, args.Routing)
}

func (r *MutationResolver) RenameRouting(args *struct {
	ID   graphql.ID
	Name string
}) (int32, error) {
	return routing.Rename(context.TODO(), args.ID, args.Name)
}

func (r *MutationResolver) RemoveRouting(args *struct {
	ID graphql.ID
}) (int32, error) {
	return routing.Remove(context.TODO(), args.ID)
}

func (r *MutationResolver) SelectRouting(args *struct {
	ID graphql.ID
}) (int32, error) {
	return routing.Select(context.TODO(), args.ID)
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
	PolicyParams *[]struct {
		Key *string
		Val string
	}
}) (*group.Resolver, error) {
	var policyParams []config_parser.Param
	if args.PolicyParams != nil {
		// Convert.
		var params []config_parser.Param
		for _, p := range *args.PolicyParams {
			var k string
			if p.Key != nil {
				k = *p.Key
			}
			params = append(params, config_parser.Param{
				Key: k,
				Val: p.Val,
			})
		}
		policyParams = params
	}
	return group.Create(context.TODO(), args.Name, args.Policy, policyParams)
}

func (r *MutationResolver) GroupSetPolicy(args *struct {
	ID           graphql.ID
	Policy       string
	PolicyParams *[]struct {
		Key *string
		Val string
	}
}) (int32, error) {
	var policyParams []config_parser.Param
	if args.PolicyParams != nil {
		// Convert.
		var params []config_parser.Param
		for _, p := range *args.PolicyParams {
			var k string
			if p.Key != nil {
				k = *p.Key
			}
			params = append(params, config_parser.Param{
				Key: k,
				Val: p.Val,
			})
		}
		policyParams = params
	}
	return group.SetPolicy(context.TODO(), args.ID, args.Policy, policyParams)
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
