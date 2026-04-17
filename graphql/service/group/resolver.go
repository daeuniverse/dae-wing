package group

import (
	"context"
	"regexp"
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

type SubscriptionBindingResolver struct {
	*db.GroupSubscription
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

func matchedNodesForBinding(binding *db.GroupSubscription) ([]db.Node, error) {
	nodes := binding.Subscription.Node
	if binding.NameFilterRegex == nil || *binding.NameFilterRegex == "" {
		return nodes, nil
	}

	re, err := regexp.Compile(*binding.NameFilterRegex)
	if err != nil {
		return nil, err
	}

	var matched []db.Node
	for _, n := range nodes {
		if re.MatchString(n.Name) {
			matched = append(matched, n)
		}
	}
	return matched, nil
}

func (r *Resolver) Subscriptions() (rs []*SubscriptionBindingResolver, err error) {
	var bindings []db.GroupSubscription
	if err = db.DB(context.TODO()).
		Where("group_id = ?", r.Group.ID).
		Preload("Subscription").
		Preload("Subscription.Node").
		Find(&bindings).Error; err != nil {
		return nil, err
	}
	for _, _binding := range bindings {
		binding := _binding
		rs = append(rs, &SubscriptionBindingResolver{GroupSubscription: &binding})
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

func (r *SubscriptionBindingResolver) Subscription() *subscription.Resolver {
	sub := r.GroupSubscription.Subscription
	return &subscription.Resolver{Subscription: &sub}
}

func (r *SubscriptionBindingResolver) NameFilterRegex() *string {
	return r.GroupSubscription.NameFilterRegex
}

func (r *SubscriptionBindingResolver) MatchedNodes() (rs []*node.Resolver, err error) {
	matched, err := matchedNodesForBinding(r.GroupSubscription)
	if err != nil {
		return nil, err
	}
	for _, _node := range matched {
		n := _node
		rs = append(rs, &node.Resolver{Node: &n})
	}
	return rs, nil
}

func (r *SubscriptionBindingResolver) MatchedCount() (int32, error) {
	matched, err := matchedNodesForBinding(r.GroupSubscription)
	if err != nil {
		return 0, err
	}
	return int32(len(matched)), nil
}
