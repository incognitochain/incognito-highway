package process

import (
	"context"
	"highway/chaindata"
	"highway/process/datahandler"
	"highway/process/topic"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

// TODO @0xakk0r0kamui remove this global param in next pull request
var GlobalPubsub PubSubManager

type SubHandler struct {
	Topic   string
	Handler func(*p2pPubSub.Subscription)
	Sub     *p2pPubSub.Subscription
	Locker  *sync.Mutex
}

type PubSubManager struct {
	SupportShards        []byte
	FloodMachine         *p2pPubSub.PubSub
	GossipMachine        *p2pPubSub.PubSub
	SubHandlers          chan SubHandler
	OutSideMessage       chan string
	followedTopic        []string
	ForwardNow           chan p2pPubSub.Message
	SpecialPublishTicker *time.Ticker
	BlockChainData       *chaindata.ChainData
}

func InitPubSub(
	s host.Host,
	supportShards []byte,
	chainData *chaindata.ChainData,
) error {
	ctx := context.Background()
	var err error
	GlobalPubsub.FloodMachine, err = p2pPubSub.NewFloodSub(ctx, s)
	if err != nil {
		return err
	}
	GlobalPubsub.SubHandlers = make(chan SubHandler, 100)
	GlobalPubsub.ForwardNow = make(chan p2pPubSub.Message)
	GlobalPubsub.SpecialPublishTicker = time.NewTicker(5 * time.Second)
	GlobalPubsub.SupportShards = supportShards
	topic.Handler.UpdateSupportShards(supportShards)
	logger.Infof("Supported shard %v", supportShards)
	GlobalPubsub.BlockChainData = chainData
	GlobalPubsub.SubKnownTopics(true)
	GlobalPubsub.SubKnownTopics(false)
	return nil
}

func (pubsub *PubSubManager) WatchingChain() {
	for {
		select {
		case newSubHandler := <-pubsub.SubHandlers:
			logger.Infof("Watching chain sub topic %v", newSubHandler.Topic)
			subch, err := pubsub.FloodMachine.Subscribe(newSubHandler.Topic)
			pubsub.followedTopic = append(pubsub.followedTopic, newSubHandler.Topic)
			if err != nil {
				logger.Info(err)
				continue
			}
			go newSubHandler.Handler(subch)
		case <-pubsub.SpecialPublishTicker.C:
			go pubsub.PublishPeerStateToNode()
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

func (pubsub *PubSubManager) SubKnownTopics(fromInside bool) error {
	var topicSubs []string
	if fromInside {
		topicSubs = topic.Handler.GetListSubTopicForHW()
	} else {
		topicSubs = topic.Handler.GetAllTopicOutsideForHW()
	}
	for _, topicSub := range topicSubs {
		subs, err := pubsub.FloodMachine.Subscribe(topicSub)
		if err != nil {
			logger.Info(err)
			continue
			// TODO retry subscribe
		}
		logger.Infof("Success subscribe topic %v", topicSub)
		pubsub.followedTopic = append(pubsub.followedTopic, topicSub)
		handler := datahandler.SubsHandler{
			PubSub:         pubsub.FloodMachine,
			FromInside:     fromInside,
			BlockchainData: pubsub.BlockChainData,
		}
		go func() {
			err := handler.HandlerNewSubs(subs)
			if err != nil {
				logger.Errorf("Handle Subsciption topic %v return error %v", subs.Topic(), err)
			}
			// TODO close subsciption and retry subscribe
		}()
	}
	return nil
}
