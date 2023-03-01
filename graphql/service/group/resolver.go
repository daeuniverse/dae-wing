package group

import (
	"context"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql/config"
	"github.com/v2rayA/dae-wing/graphql/service/node"
	"github.com/v2rayA/dae-wing/graphql/service/subscription"
)

type Resolver struct {
	*db.Group
}

func (r *Resolver) Name() string {
	return r.Group.Name
}

func (r *Resolver) Nodes() (rs []*node.Resolver) {
	var m db.Group
	db.DB(context.TODO()).Model(r).Preload("Node").Find(&m)
	for _, _n := range r.Group.Node {
		n := _n
		rs = append(rs, &node.Resolver{Node: &n})
	}
	return rs
}

func (r *Resolver) Subscriptions() (rs []*subscription.Resolver) {
	var m db.Group
	db.DB(context.TODO()).Model(r).Preload("Subscription").Find(&m)
	for _, _n := range m.Subscription {
		n := _n
		rs = append(rs, &subscription.Resolver{Subscription: &n})
	}
	return rs
}

func (r *Resolver) Policy() *config.AndFunctionsOrPlaintextResolver {
	return &config.AndFunctionsOrPlaintextResolver{
		FunctionListOrString: r.Group.Policy,
	}
}
