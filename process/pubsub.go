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
	Handler func(*p2pPubSub.Message)
}

// type Config struct {
// 	Suppo
// }

type PubSubManager struct {
	SupportShards        []byte
	FloodMachine         *p2pPubSub.PubSub
	GossipMachine        *p2pPubSub.PubSub
	GRPCMessage          chan string
	OutSideMessage       chan string
	followedTopic        []string
	outsideMessage       []string
	ForwardNow           chan p2pPubSub.Message
	Msgs                 []*p2pPubSub.Subscription
	SpecialPublishTicker *time.Ticker
	BlockChainData       *ChainData
}

func InitPubSub(
	s host.Host,
	supportShards []byte,
	chainData *ChainData,
) error {
	ctx := context.Background()
	var err error
	GlobalPubsub.FloodMachine, err = p2pPubSub.NewFloodSub(ctx, s)
	if err != nil {
		return err
	}
	GlobalPubsub.GRPCMessage = make(chan string)
	GlobalPubsub.ForwardNow = make(chan p2pPubSub.Message)
	GlobalPubsub.Msgs = make([]*p2pPubSub.Subscription, 0)
	GlobalPubsub.SpecialPublishTicker = time.NewTicker(5 * time.Second)
	GlobalPubsub.SupportShards = supportShards
	GlobalPubsub.BlockChainData = chainData
	// topic.InitTypeOfProcessor()
	// initGlobalParams()
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
			logger.Infof("Topic %v, Type of processor %v", newTopic, typeOfProcessor)
			logger.Infof("Success subscribe topic %v, Type of process %v", newTopic, typeOfProcessor)
			pubsub.Msgs = append(pubsub.Msgs, subch)
			go pubsub.handleNewMsg(subch, typeOfProcessor)
		case <-pubsub.SpecialPublishTicker.C:
			for committeeID, committeeState := range pubsub.BlockChainData.ListMsgPeerStateOfShard {
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
		sub.Cancel()
		//TODO implement GossipSub with special topic
		if (err == nil) && (data != nil) {
			switch typeOfProcessor {
			case topic.DoNothing:
				continue
			case topic.ProcessAndPublishAfter:
				// if topic.GetMsgTypeOfTopic(sub.Topic()) == topic.CmdPeerState {
				// 	logger.Infof("Receive data from topic Peer state")
				// 	x, err := ParsePeerStateData(string(data.GetData()))
				// 	if err != nil {
				// 		logger.Error(err)
				// 	} else {
				// 		logger.Infof("PeerState data:\n Beacon: %v\n Shard: %v\n", x.Beacon, x.Shards)
				// 	}
				// }
				// logger.Infof("Receive data ProcessAndPublishAfter")
				go pubsub.BlockChainData.UpdatePeerState(pubsub.BlockChainData.CommitteePubkeyByPeerID[data.GetFrom()], data.GetData())
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
