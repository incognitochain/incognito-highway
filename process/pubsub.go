package process

import (
	"context"
	"highway/database"
	"highway/process/topic"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
)

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
	GlobalPubsub.SubHandlers = make(chan SubHandler, 100)
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

func (handler *SubHandler) handleNewMsgv2(
	pubsub *PubSubManager,
	fromInside bool,
) {
	for {
		data, err := handler.Sub.Next(context.Background())
		dataBytes := data.GetData()
		if (err == nil) && (data != nil) {

			//#region check duplicate and cache data
			handler.Locker.Lock()
			if database.IsMarkedData(dataBytes) {
				handler.Locker.Unlock()
				continue
			}
			database.MarkData(dataBytes)
			handler.Locker.Unlock()
			//#endregion check duplicate and cache data

			msgType := topic.GetMsgTypeOfTopic(handler.Topic)
			cID := topic.GetCommitteeIDOfTopic(handler.Topic)
			// TODO @0xakk0r0kamui replace it by map[string]func
			switch handler.Topic {
			case topic.CmdBFT:
				continue
			case topic.CmdPeerState:
				go pubsub.HandlePeerState(data.GetFrom(), dataBytes)
			case topic.CmdBlockBeacon:
				go pubsub.HandleBlkBeacon(fromInside, msgType, cID, dataBytes)
			case topic.CmdBlkShardToBeacon, topic.CmdCrossShard:
				go pubsub.HandleCrossCommitteeBlock(fromInside, msgType, cID, dataBytes)
			case topic.CmdBlockShard:
				go pubsub.HandleBlkShard(msgType, cID, dataBytes)
			case topic.CmdTx, topic.CmdPrivacyCustomToken, topic.CmdCustomToken:
				go pubsub.HandleTX(msgType, cID, dataBytes)
			default:
				logger.Warnf("Missing handler for topic %v", handler.Topic)
				return
			}
		}
	}
}

func (pubsub *PubSubManager) handleNewMsg(
	sub *p2pPubSub.Subscription,
	typeOfProcessor byte,
) {
	for {
		isDuplicate := false
		data, err := sub.Next(context.Background())
		// TODO implement GossipSub with special topic
		// TODO Add lock for each of msg type
		cID := topic.GetCommitteeIDOfTopic(sub.Topic())
		//Just temp fix, updated in 1 HW 1 Shard version
		cIDByte := byte(cID)
		if cID == -1 {
			cIDByte = 254
		}
		data4cache := append(data.GetData(), cIDByte)
		if (err == nil) && (data != nil) {
			if database.IsMarkedData(data4cache) {
				isDuplicate = true
			} else {
				database.MarkData(data4cache)
			}
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
				go func() {
					err := pubsub.BlockChainData.UpdatePeerState(pubsub.BlockChainData.GetMiningPubkeyFromPeerID(data.GetFrom()), data.GetData())
					if err != nil {
						err = errors.WithMessagef(err, "from: %v", data.GetFrom())
						logger.Warnf("Cannot update peerstate: %+v", err)
					}
				}()

			case topic.ProcessAndPublish:
				logger.Debugf("[pubsub] Received data of topic %v, data [%v..%v]", sub.Topic(), data.Data[:5], data.Data[len(data.Data)-6:])
				if isDuplicate {
					continue
				}
				logger.Debugf("[pubsub] Broadcast topic %v, data [%v..%v]", sub.Topic(), data.Data[:5], data.Data[len(data.Data)-6:])
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
		logger.Infof("Success subscribe topic %v", topicSub)
		subch, err := pubsub.FloodMachine.Subscribe(topicSub)
		pubsub.followedTopic = append(pubsub.followedTopic, topicSub)
		if err != nil {
			logger.Info(err)
			continue
		}
		typeOfProcessor := topic.GetTypeOfProcess(topicSub)
		go pubsub.handleNewMsg(subch, typeOfProcessor)
	}
	return nil
}
