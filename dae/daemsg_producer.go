package dae

import (
	"container/list"
	"context"

	"github.com/daeuniverse/dae/control"
)

type DaeMsgProducer struct {
	chMsg <-chan *control.Msg
	// resetChMsg is used to re-assign chMsg.
	resetChMsg    chan struct{}
	daeSubscriber chan *daeSubscriber
}

func NewDaeMsgProducer(chMsg <-chan *control.Msg) *DaeMsgProducer {
	r := &DaeMsgProducer{
		chMsg:         chMsg,
		resetChMsg:    make(chan struct{}),
		daeSubscriber: make(chan *daeSubscriber),
	}
	return r
}

func (r *DaeMsgProducer) ReassignChMsg(chMsg <-chan *control.Msg) {
	r.chMsg = chMsg
	r.resetChMsg <- struct{}{}
}

func (r *DaeMsgProducer) Run() {
	subs := list.New().Init()
	for {
		select {
		// New subscriber.
		case sub := <-r.daeSubscriber:
			subs.PushBack(sub)
		// Reset pointer to r.chMsg.
		case <-r.resetChMsg:

		// Producer msg.
		case msg := <-r.chMsg:
			for node := subs.Front(); node != nil; node = node.Next() {
				sub := node.Value.(*daeSubscriber)
				select {
				case <-sub.stop:
					subs.Remove(node)
				case sub.msgs <- msg:
				default:
					// Busy.
				}
			}
		}
	}
}

type daeSubscriber struct {
	msgs chan<- *control.Msg
	stop <-chan struct{}
}

func (r *DaeMsgProducer) Subscribe(ctx context.Context) <-chan *control.Msg {
	c := make(chan *control.Msg, 10)
	r.daeSubscriber <- &daeSubscriber{msgs: c, stop: ctx.Done()}
	return c
}
