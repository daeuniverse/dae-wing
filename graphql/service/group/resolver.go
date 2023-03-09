package group

import (
	"context"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/daeuniverse/dae-wing/graphql/service/subscription"
	"github.com/graph-gophers/graphql-go"
)

type Resolver struct {
	*db.Group
}

func (r *Resolver) ID() graphql.ID {
	return common.EncodeCursor(r.Group.ID)
}

func (r *Resolver) Name() string {
	return r.Group.Name
}

func (r *Resolver) Nodes() (rs []*node.Resolver, err error) {
	var nodes []db.Node
	if err = db.DB(context.TODO()).Model(r.Group).Association("Node").Find(&nodes); err != nil {
		return nil, err
	}
	for _, _n := range nodes {
		n := _n
		rs = append(rs, &node.Resolver{Node: &n})
	}
	return rs, nil
}

func (r *Resolver) Subscriptions() (rs []*subscription.Resolver, err error) {
	var subs []db.Subscription
	if err = db.DB(context.TODO()).Model(r.Group).Association("Subscription").Find(&subs); err != nil {
		return nil, err
	}
	for _, _n := range subs {
		n := _n
		rs = append(rs, &subscription.Resolver{Subscription: &n})
	}
	return rs, nil
}

func (r *Resolver) Policy() string {
	return r.Group.Policy
}

func (r *Resolver) PolicyParams() (rs []*internal.ParamResolver, err error) {
	var params []db.GroupPolicyParam
	if err = db.DB(context.TODO()).Model(r.Group).Association("PolicyParams").Find(&params); err != nil {
		return nil, err
	}
	for _, param := range params {
		rs = append(rs, &internal.ParamResolver{Param: param.Marshal()})
	}
	return rs, nil
}
