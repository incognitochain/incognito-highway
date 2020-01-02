package process

import (
	"context"
	"highway/chaindata"
	"highway/common"
	"highway/process/datahandler"
	"highway/process/topic"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/patrickmn/go-cache"
)

// TODO @0xakk0r0kamui remove this global param in next pull request
// var GlobalPubsub PubSubManager

type SubHandler struct {
	Topic   string
	Handler func(*p2pPubSub.Subscription)
	Sub     *p2pPubSub.Subscription
	Locker  *sync.Mutex
}

type PubSubManager struct {
	SupportShards        []byte
	FloodMachine         *p2pPubSub.PubSub
	SubHandlers          chan SubHandler
	followedTopic        []string
	SpecialPublishTicker *time.Ticker
	BlockChainData       *chaindata.ChainData
}

func NewPubSub(
	s host.Host,
	supportShards []byte,
	chainData *chaindata.ChainData,
) (
	*PubSubManager,
	error,
) {
	pubsub := new(PubSubManager)
	ctx := context.Background()
	var err error
	pubsub.FloodMachine, err = p2pPubSub.NewFloodSub(ctx, s)
	if err != nil {
		return nil, err
	}
	pubsub.SubHandlers = make(chan SubHandler, 100)
	pubsub.SpecialPublishTicker = time.NewTicker(common.TimeIntervalPublishStates)
	pubsub.SupportShards = supportShards
	pubsub.BlockChainData = chainData
	// if cacher == nil {
	// 	pubsub.Cacher = cache.New(common.MaxTimeKeepPubSubData, common.MaxTimeKeepPubSubData)
	// } else {
	// 	pubsub.Cacher = cacher
	// }
	// topic.Handler.UpdateSupportShards(supportShards)
	logger.Infof("Supported shard %v", supportShards)
	err = pubsub.SubKnownTopics(true)
	if err != nil {
		logger.Errorf("Subscribe topic from node return error %v", err)
		return nil, err
	}
	err = pubsub.SubKnownTopics(false)
	if err != nil {
		logger.Errorf("Subscribe topic from other HWs return error %v", err)
		return nil, err
	}
	return pubsub, nil
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
	cacher := cache.New(common.MaxTimeKeepPubSubData, common.MaxTimeKeepPubSubData)
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
			Cacher:         cacher,
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

// func deletePeerIDinSlice(target peer.ID, slice []peer.ID) []peer.ID {
// 	i := 0
// 	for _, peerID := range slice {
// 		if peerID.String() != target.String() {
// 			slice[i] = peerID
// 			i++
// 		}
// 	}
// 	return slice[:i]
// }

// For Grafana, will complete in next pull request
// func (pubsub *PubSubManager) WatchingSubs(subs *p2pPubSub.Subscription) {
// 	for {
// 		event, err := subs.NextPeerEvent(context.Background())
// 		if err != nil {
// 			logger.Error(err)
// 		}
// 		if event.Type == p2pPubSub.PeerJoin {
// 			pubsub.pubsubInfo.Locker.Lock()
// 			pubsub.pubsubInfo.Info[subs.Topic()] = append(pubsub.pubsubInfo.Info[subs.Topic()], event.Peer)
// 			pubsub.pubsubInfo.Locker.Unlock()
// 		}
// 		if event.Type == p2pPubSub.PeerLeave {
// 			pubsub.pubsubInfo.Locker.Lock()
// 			pubsub.pubsubInfo.Info[subs.Topic()] = deletePeerIDinSlice(event.Peer, pubsub.pubsubInfo.Info[subs.Topic()])
// 			pubsub.pubsubInfo.Locker.Unlock()
// 		}
// 	}
// }
