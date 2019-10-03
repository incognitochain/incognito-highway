package process

import (
	"context"
	fmt "fmt"
	logger "highway/customizelog"

	"github.com/libp2p/go-libp2p-core/host"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

var GlobalPubsub PubSubManager

type SubHandler struct {
	Topic   string
	Handler func(*p2pPubSub.Message)
}

type PubSubManager struct {
	FloodMachine   *p2pPubSub.PubSub
	GossipMachine  *p2pPubSub.PubSub
	NewMessage     chan SubHandler
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
	// GlobalPubsub.GossipMachine, err = p2pPubSub.NewGossipSub(ctx, s)
	// if err != nil {
	// 	return err
	// }
	GlobalPubsub.NewMessage = make(chan SubHandler, 100)
	GlobalPubsub.ForwardNow = make(chan p2pPubSub.Message)
	GlobalPubsub.Msgs = make([]*p2pPubSub.Subscription, 0)
	return nil
}

func (pubsub *PubSubManager) WatchingChain() {
	for {
		select {
		case subHandler := <-pubsub.NewMessage:
			subch, err := pubsub.FloodMachine.Subscribe(subHandler.Topic)
			if err != nil {
				logger.Info(err)
				continue
			}
			logger.Infof("Success subscribe topic %v\n", subHandler.Topic)
			pubsub.Msgs = append(pubsub.Msgs, subch)
			go pubsub.handleNewMsg(subch, subHandler.Handler)
		}
	}
}

func (pubsub *PubSubManager) handleNewMsg(sub *p2pPubSub.Subscription, handler func(*p2pPubSub.Message)) {
	for {
		data, err := sub.Next(context.Background())
		fmt.Println("~~~~~~~~~~", err, "~~~~~~~~~~", data, "~~~~~~~~~~")
		//TODO implement GossipSub with special topic
		if (err == nil) && (data != nil) {
			// err = pubsub.FloodMachine.Publish(sub.Topic(), data.GetData())
			if err == nil {
				logger.Infof("Topic: %v, data: %v\n", sub.Topic(), string(data.Data))
			} else {
				logger.Infof("Publish topic %v failed, err: %v\n", sub.Topic(), err)
			}
			handler(data)
		}
	}
}
