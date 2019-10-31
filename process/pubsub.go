package process

import (
	"context"
	logger "highway/customizelog"
	"highway/process/topic"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

var GlobalPubsub PubSubManager

type SubHandler struct {
	Topic   string
	Handler func(*p2pPubSub.Subscription)
}

// type Config struct {
// 	Suppo
// }

type PubSubManager struct {
	SupportShards        []byte
	FloodMachine         *p2pPubSub.PubSub
	GossipMachine        *p2pPubSub.PubSub
	GRPCMessage          chan string
	GRPCSpecSub          chan SubHandler
	OutSideMessage       chan string
	followedTopic        []string
	outsideMessage       []string
	ForwardNow           chan p2pPubSub.Message
	Msgs                 []*p2pPubSub.Subscription
	SpecialPublishTicker *time.Ticker
}

func InitPubSub(s host.Host, supportShards []byte) error {
	ctx := context.Background()
	var err error
	GlobalPubsub.FloodMachine, err = p2pPubSub.NewFloodSub(ctx, s)
	if err != nil {
		return err
	}
	GlobalPubsub.GRPCMessage = make(chan string)
	GlobalPubsub.GRPCSpecSub = make(chan SubHandler, 100)
	GlobalPubsub.ForwardNow = make(chan p2pPubSub.Message)
	GlobalPubsub.Msgs = make([]*p2pPubSub.Subscription, 0)
	GlobalPubsub.SpecialPublishTicker = time.NewTicker(5 * time.Second)
	GlobalPubsub.SupportShards = supportShards
	topic.InitTypeOfProcessor()
	// TODO hy remove global param
	initGlobalParams()
	return nil
}

func (pubsub *PubSubManager) WatchingChain() {
	for {
		select {
		case newTopic := <-pubsub.GRPCMessage:
			subch, err := pubsub.FloodMachine.Subscribe(newTopic)
			pubsub.followedTopic = append(pubsub.followedTopic, newTopic)
			if err != nil {
				logger.Info(err)
				continue
			}
			typeOfProcessor := topic.GetTypeOfProcess(newTopic)
			// logger.Infof("Topic %v, Type of processor %v", newTopic, typeOfProcessor)

			// logger.Infof("Success subscribe topic %v, Type of process %v", newTopic, typeOfProcessor)
			pubsub.Msgs = append(pubsub.Msgs, subch)
			go pubsub.handleNewMsg(subch, typeOfProcessor)
		case newGRPCSpecSub := <-pubsub.GRPCSpecSub:
			subch, err := pubsub.FloodMachine.Subscribe(newGRPCSpecSub.Topic)
			pubsub.followedTopic = append(pubsub.followedTopic, newGRPCSpecSub.Topic)
			if err != nil {
				logger.Info(err)
				continue
			}
			// typeOfProcessor := topic.GetTypeOfProcess(newTopic)
			// logger.Infof("Received new special sub from GRPC, topic: %v", newGRPCSpecSub.Topic)

			// logger.Infof("Success subscribe topic %v, Type of process %v", newTopic, typeOfProcessor)
			// pubsub.Msgs = append(pubsub.Msgs, subch)
			go newGRPCSpecSub.Handler(subch)
		case <-pubsub.SpecialPublishTicker.C:
			for committeeID, committeeState := range AllPeerState {
				for _, stateData := range committeeState {
					PeriodicalPublish(pubsub.FloodMachine, topic.CmdPeerState, committeeID, stateData)
				}
			}
		}

	}
}

func (pubsub *PubSubManager) handleNewMsg(sub *p2pPubSub.Subscription, typeOfProcessor byte) {
	for {
		data, err := sub.Next(context.Background())
		//TODO implement GossipSub with special topic
		if (err == nil) && (data != nil) {
			switch typeOfProcessor {
			case topic.DoNothing:
				// logger.Infof("Receive data from topic %v DoNothing", sub.Topic())
				continue
			case topic.ProcessAndPublishAfter:
				// logger.Infof("Receive data ProcessAndPublishAfter")
				go UpdatePeerState(CommitteePubkeyByPeerID[data.GetFrom()], data.GetData())
			case topic.ProcessAndPublish:
				//TODO hy add handler(data)
				go ProcessNPublishDataFromTopic(pubsub.FloodMachine, sub.Topic(), data.GetData(), pubsub.SupportShards)
			default:
				return
			}
			// handler(data)
		}
	}
}

func (pubsub *PubSubManager) hasTopic(receivedTopic string) bool {
	for _, flTopic := range pubsub.followedTopic {
		if receivedTopic == flTopic {
			return true
		}
	}
	return false
}
