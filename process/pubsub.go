package process

import (
	"context"
	"fmt"
	"highway/chaindata"
	"highway/common"
	"highway/grafana"
	"highway/process/datahandler"
	"highway/process/simulateutils"
	"highway/process/topic"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-peer"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
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
	FollowedTopic        []string
	SpecialPublishTicker *time.Ticker
	BlockChainData       *chaindata.ChainData
	Info                 *PubsubInfo
	GraLog               *grafana.GrafanaLog
	CommitteeInfo        *simulateutils.CommitteeTable
	Scenario             *simulateutils.Scenario
}

func NewPubSub(
	s host.Host,
	supportShards []byte,
	chainData *chaindata.ChainData,
	scenario *simulateutils.Scenario,
	cInfo *simulateutils.CommitteeTable,
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
	pubsub.CommitteeInfo = cInfo
	pubsub.Scenario = scenario
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
			pubsub.FollowedTopic = append(pubsub.FollowedTopic, newSubHandler.Topic)
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
	for _, flTopic := range pubsub.FollowedTopic {
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
		pubsub.FollowedTopic = append(pubsub.FollowedTopic, topicSub)
		handler := datahandler.SubsHandler{
			PubSub:         pubsub.FloodMachine,
			FromInside:     fromInside,
			BlockchainData: pubsub.BlockChainData,
			Cacher:         cacher,
			CommitteeInfo:  pubsub.CommitteeInfo,
			Scenario:       pubsub.Scenario,
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
func (pubsub *PubSubManager) WatchingSubs(subs *p2pPubSub.Subscription) {
	for {
		event, err := subs.NextPeerEvent(context.Background())
		if err != nil {
			logger.Error(err)
		}
		tp := subs.Topic()
		msgType := topic.GetMsgTypeOfTopic(tp)
		if event.Type == p2pPubSub.PeerJoin {
			pubsub.Info.Locker.Lock()
			pubsub.Info.Info[tp] = append(pubsub.Info.Info[tp], event.Peer)
			if pubsub.GraLog != nil {
				pubsub.GraLog.Add(fmt.Sprintf("total_%s", msgType), len(pubsub.Info.Info[tp]))
			}
			pubsub.Info.Locker.Unlock()
		}
		if event.Type == p2pPubSub.PeerLeave {
			pubsub.Info.Locker.Lock()
			pubsub.Info.Info[subs.Topic()] = deletePeerIDinSlice(event.Peer, pubsub.Info.Info[subs.Topic()])
			if pubsub.GraLog != nil {
				pubsub.GraLog.Add(fmt.Sprintf("total_%s", msgType), len(pubsub.Info.Info[tp]))
			}
			pubsub.Info.Locker.Unlock()
		}
	}
}
