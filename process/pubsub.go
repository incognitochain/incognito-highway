package process

import (
	"context"
	fmt "fmt"
	logger "highway/customizelog"

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
	GlobalPubsub.NewMessage = make(chan string)
	GlobalPubsub.ForwardNow = make(chan p2pPubSub.Message)
	GlobalPubsub.Msgs = make([]*p2pPubSub.Subscription, 0)
	return nil
}

func (pubsub *PubSubManager) WatchingChain() {
	for {
		select {
		case newTopic := <-pubsub.NewMessage:
			subch, err := pubsub.FloodMachine.Subscribe(newTopic)
			if err != nil {
				logger.Info(err)
				continue
			}
			logger.Infof("Success subscribe topic %v\n", newTopic)
			pubsub.Msgs = append(pubsub.Msgs, subch)
			go pubsub.handleNewMsg(subch)
		}
	}
}

func (pubsub *PubSubManager) handleNewMsg(sub *p2pPubSub.Subscription) {
	for {
		data, err := sub.Next(context.Background())
		fmt.Println("~~~~~~~~~~", err, "~~~~~~~~~~", data, "~~~~~~~~~~")
		//TODO implement GossipSub with special topic
		if (err == nil) && (data != nil) {
			err = pubsub.FloodMachine.Publish(sub.Topic(), data.GetData())
			if err == nil {
				logger.Infof("Success publish topic %v\n")
				logger.Infof("Topic: %v, data: %v\n", sub.Topic(), data.Data)
			} else {
				logger.Infof("Publish topic %v failed, err: %v\n", sub.Topic(), err)
			}
		}
	}
}
