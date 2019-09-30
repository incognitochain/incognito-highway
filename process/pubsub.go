package process

import (
	"context"
	fmt "fmt"

	"github.com/libp2p/go-libp2p-core/host"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

var GlobalPubsub PubSubManager

type PubSubManager struct {
	FloodMachine   *p2pPubSub.PubSub
	GossipMachine  *p2pPubSub.PubSub
	NewMessage     chan string
	insideMessage  []string
	outsideMessage []string
	ForwardNow     chan p2pPubSub.Message
	Msgs           []*p2pPubSub.Subscription
}

func InitPubSub(s host.Host) error {
	ctx := context.Background()
	var err error
	GlobalPubsub.FloodMachine, err = p2pPubSub.NewFloodSub(ctx, s)
	if err != nil {
		return err
	}
	GlobalPubsub.GossipMachine, err = p2pPubSub.NewGossipSub(ctx, s)
	if err != nil {
		return err
	}
	GlobalPubsub.Msgs = make([]*p2pPubSub.Subscription, 0)
	return nil
}

func (pubsub *PubSubManager) WatchingChain() {
	for {
		select {
		case newTopic := <-pubsub.NewMessage:
			subch, err := pubsub.FloodMachine.Subscribe(newTopic)
			if err != nil {
				fmt.Println(err)
			}
			pubsub.Msgs = append(pubsub.Msgs, subch)
			go pubsub.handleNewMess(subch)
		}
	}
}

func (pubsub *PubSubManager) handleNewMess(x *p2pPubSub.Subscription) {
	a, err := x.Next(nil)
	//TODO implement GossipSub with special topic
	if (err == nil) && (a != nil) {
		pubsub.FloodMachine.Publish(x.Topic(), a.GetData())
	}
}
