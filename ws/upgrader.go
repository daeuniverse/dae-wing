package ws

import (
	"context"
	"sync"
	"time"

	"github.com/daeuniverse/dae-wing/dae"
	jsoniter "github.com/json-iterator/go"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var upgrader = newUpgrader()

func subscribe(ctx context.Context, mp *dae.DaeMsgProducer, c *websocket.Conn) {
	chMsg := mp.Subscribe(ctx)
	go func() {
		for msg := range chMsg {
			bMsg, err := jsoniter.Marshal(msg)
			if err != nil {
				panic(err)
			}
			c.WriteMessage(websocket.TextMessage, bMsg)
		}
	}()
}

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	type Subscripber struct {
		cancel context.CancelFunc
	}
	subscribers := sync.Map{}
	u.OnOpen(func(c *websocket.Conn) {
		identifier := c.RemoteAddr().String()
		ctx, cancel := context.WithCancel(context.TODO())
		subscribers.Store(identifier, &Subscripber{
			cancel: cancel,
		})
		if mp := dae.MsgProducer; mp != nil {
			subscribe(ctx, mp, c)
		} else {
			// Waiting for the initializing of dae.MsgProducer.
			go func(ctx context.Context) {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()
				for range ticker.C {
					select {
					case <-ctx.Done():
						// User has left.
						return
					default:
						if mp := dae.MsgProducer; mp != nil {
							// dae.MsgProducer is initialized.
							subscribe(ctx, mp, c)
						}
					}
				}
			}(ctx)
		}
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		identifier := c.RemoteAddr().String()
		subscriber, ok := subscribers.LoadAndDelete(identifier)
		if ok {
			subscriber.(*Subscripber).cancel()
		}
	})
	return u
}
