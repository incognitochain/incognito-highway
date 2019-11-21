package process

import (
	"context"
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

type PubSubManager struct {
	SupportShards        []byte
	FloodMachine         *p2pPubSub.PubSub
	GossipMachine        *p2pPubSub.PubSub
	GRPCSpecSub          chan SubHandler
	OutSideMessage       chan string
	followedTopic        []string
	ForwardNow           chan p2pPubSub.Message
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
	GlobalPubsub.GRPCSpecSub = make(chan SubHandler, 100)
	GlobalPubsub.ForwardNow = make(chan p2pPubSub.Message)
	GlobalPubsub.SpecialPublishTicker = time.NewTicker(5 * time.Second)
	GlobalPubsub.SupportShards = supportShards
	GlobalPubsub.BlockChainData = chainData
	topic.InitTypeOfProcessor()
	GlobalPubsub.SubKnownTopics()
	return nil
}

func (pubsub *PubSubManager) WatchingChain() {
	for {
		select {
		case newGRPCSpecSub := <-pubsub.GRPCSpecSub:
			subch, err := pubsub.FloodMachine.Subscribe(newGRPCSpecSub.Topic)
			pubsub.followedTopic = append(pubsub.followedTopic, newGRPCSpecSub.Topic)
			if err != nil {
				logger.Info(err)
				continue
			}
			go newGRPCSpecSub.Handler(subch)
		case <-pubsub.SpecialPublishTicker.C:
			go pubsub.PublishPeerStateToNode()
		}

	}
}

func (pubsub *PubSubManager) handleNewMsg(
	sub *p2pPubSub.Subscription,
	typeOfProcessor byte,
) {
	for {
		data, err := sub.Next(context.Background())
		//TODO implement GossipSub with special topic
		if (err == nil) && (data != nil) {
			switch typeOfProcessor {
			case topic.DoNothing:
				continue
			case topic.ProcessAndPublishAfter:
				//#region Just logging information
				// if topic.GetMsgTypeOfTopic(sub.Topic()) == topic.CmdPeerState {
				// 	x, err := ParsePeerStateData(string(data.GetData()))
				// 	if err != nil {
				// 		logger.Error(err)
				// 	} else {
				// 		logger.Infof("PeerState data:\n Beacon: %v\n Shard: %v\n", x.Beacon, x.Shards)
				// 	}
				// }
				//#endregion Just logging information
				go pubsub.BlockChainData.UpdatePeerState(pubsub.BlockChainData.CommitteePubkeyByPeerID[data.GetFrom()], data.GetData())
			case topic.ProcessAndPublish:
				pubTopics := topic.Handler.GetHWPubTopicsFromHWSub(sub.Topic())
				for _, pubTopic := range pubTopics {
					go pubsub.FloodMachine.Publish(pubTopic, data.GetData())
				}
			default:
				return
			}
		}
	}
}

func (pubsub *PubSubManager) HasTopic(receivedTopic string) bool {
	for _, flTopic := range pubsub.followedTopic {
		if receivedTopic == flTopic {
			return true
		}
	}
	return false
}

func (pubsub *PubSubManager) PublishPeerStateToNode() {
	for _, cID := range pubsub.SupportShards {
		pubTopics := topic.Handler.GetHWPubTopicsFromMsg(topic.CmdPeerState, int(cID))
		pubsub.BlockChainData.Locker.RLock()
		for _, stateData := range pubsub.BlockChainData.ListMsgPeerStateOfShard[cID] {
			for _, pubTopic := range pubTopics {
				err := pubsub.FloodMachine.Publish(pubTopic, stateData)
				if err != nil {
					logger.Errorf("Publish Peer state to Committee %v return error %v", cID, err)
				}
			}
		}
		pubsub.BlockChainData.Locker.RUnlock()
	}
}

func (pubsub *PubSubManager) SubKnownTopics() error {
	topicSubs := topic.Handler.GetListSubTopicForHW()
	for _, topicSub := range topicSubs {
		subch, err := pubsub.FloodMachine.Subscribe(topicSub)
		pubsub.followedTopic = append(pubsub.followedTopic, topicSub)
		if err != nil {
			logger.Info(err)
			continue
		}
		typeOfProcessor := topic.GetTypeOfProcess(topicSub)
		logger.Infof("Success subscribe topic %v, Type of process %v", topicSub, typeOfProcessor)
		go pubsub.handleNewMsg(subch, typeOfProcessor)
	}
	return nil
}
