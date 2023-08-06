package graphql

import (
	"container/list"
	"context"
	"time"

	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/graphql/service/daemsg"
)

type SubscriptionWsResolver struct {
	daeSubscriber chan *daeSubscriber
}

func newSubscriptionWsResolver() *SubscriptionWsResolver {
	r := &SubscriptionWsResolver{
		daeSubscriber: make(chan *daeSubscriber),
	}
	go r.broadcast()
	return r
}

func (r *SubscriptionWsResolver) broadcast() {
	subs := list.New().Init()
	for {
		select {
		case sub := <-r.daeSubscriber:
			subs.PushBack(sub)
		case <-dae.ResetChMsg:
		case msg := <-dae.ChMsg:
			for node := subs.Front(); node != nil; node = node.Next() {
				sub := node.Value.(*daeSubscriber)
				select {
				case <-sub.stop:
					subs.Remove(node)
				case sub.msgs <- &daemsg.Resolver{Msg: msg}:
				case <-time.After(time.Second):
				}
			}
		}
	}
}

type daeSubscriber struct {
	msgs chan<- *daemsg.Resolver
	stop <-chan struct{}
}

func (r *SubscriptionWsResolver) DaeMsg(ctx context.Context) <-chan *daemsg.Resolver {
	c := make(chan *daemsg.Resolver, 10)
	r.daeSubscriber <- &daeSubscriber{msgs: c, stop: ctx.Done()}
	return c
}
