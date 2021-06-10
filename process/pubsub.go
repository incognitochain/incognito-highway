package process

import (
	"context"
	"highway/chaindata"
	"highway/common"
	"highway/grafana"
	"highway/process/datahandler"
	"highway/process/topic"
	"sync"
	"time"

	p2pPubSub "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/patrickmn/go-cache"
)

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
	Info                 *PubsubInfo
	GraLog               *grafana.GrafanaLog
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
	// id, err := peer.Decode("Qmanj8hF7toLni73idcW1e9mNjjzBcRoQqjNWQSMPMSZEK")
	// addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/17339")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// tracer, err := p2pPubSub.NewRemoteTracer(ctx, s, peer.AddrInfo{ID: id, Addrs: []multiaddr.Multiaddr{addr}})
	pubsub.FloodMachine, err = p2pPubSub.NewFloodSub(
		ctx,
		s,
		p2pPubSub.WithMaxMessageSize(common.MaxPSMsgSize),
		p2pPubSub.WithPeerOutboundQueueSize(1024),
		p2pPubSub.WithValidateQueueSize(1024),
	)
	if err != nil {
		return nil, err
	}
	pubsub.SubHandlers = make(chan SubHandler, 100)
	pubsub.SpecialPublishTicker = time.NewTicker(common.TimeIntervalPublishStates)
	pubsub.SupportShards = supportShards
	pubsub.BlockChainData = chainData
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
		}()
	}
	return nil
}

func deletePeerIDinSlice(target peer.ID, slice []peer.ID) []peer.ID {
	i := 0
	for _, peerID := range slice {
		if peerID.String() != target.String() {
			slice[i] = peerID
			i++
		}
	}
	return slice[:i]
}

// For Grafana, will complete in next pull request
// func (pubsub *PubSubManager) WatchingSubs(subs *p2pPubSub.Subscription) {
// 	for {
// 		event, err := subs.Next() PeerEvent(context.Background())
// 		if err != nil {
// 			logger.Error(err)
// 		}
// 		tp := subs.Topic()
// 		msgType := topic.GetMsgTypeOfTopic(tp)
// 		if event.Type == p2pPubSub.PeerJoin {
// 			pubsub.Info.Locker.Lock()
// 			pubsub.Info.Info[tp] = append(pubsub.Info.Info[tp], event.Peer)
// 			if pubsub.GraLog != nil {
// 				pubsub.GraLog.Add(fmt.Sprintf("total_%s", msgType), len(pubsub.Info.Info[tp]))
// 			}
// 			pubsub.Info.Locker.Unlock()
// 		}
// 		if event.Type == p2pPubSub.PeerLeave {
// 			pubsub.Info.Locker.Lock()
// 			pubsub.Info.Info[subs.Topic()] = deletePeerIDinSlice(event.Peer, pubsub.Info.Info[subs.Topic()])
// 			if pubsub.GraLog != nil {
// 				pubsub.GraLog.Add(fmt.Sprintf("total_%s", msgType), len(pubsub.Info.Info[tp]))
// 			}
// 			pubsub.Info.Locker.Unlock()
// 		}
// 	}
// }
