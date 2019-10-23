package process

import (
	"context"
	"highway/common"
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

type PubSubManager struct {
	SupportShards        []byte
	FloodMachine         *p2pPubSub.PubSub
	GossipMachine        *p2pPubSub.PubSub
	GRPCMessage          chan string
	OutSideMessage       chan string
	followedTopic        []string
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
	topic.InitTypeOfProcessor()
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
				if topic.GetMsgTypeOfTopic(sub.Topic()) == topic.CmdPeerState {
					x, err := ParsePeerStateData(string(data.GetData()))
					if err != nil {
						logger.Error(err)
					} else {
						logger.Infof("PeerState data:\n Beacon: %v\n Shard: %v\n", x.Beacon, x.Shards)
					}
				}
				//#endregion Just logging information
				go pubsub.BlockChainData.UpdatePeerState(pubsub.BlockChainData.CommitteePubkeyByPeerID[data.GetFrom()], data.GetData())
			case topic.ProcessAndPublish:
				listPubTopic, err := pubsub.genPubTopicFromReceivedTopic(sub.Topic())
				if err != nil {
					logger.Error(err)
				} else {
					mode := OneTopicOneData
					if len(listPubTopic) > 1 {
						mode = NTopicOneData
					}
					go PublishDataWithTopic(pubsub.FloodMachine, listPubTopic, [][]byte{data.GetData()}, mode)
				}
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

func (pubsub *PubSubManager) genPubTopicFromReceivedTopic(topicReceived string) (
	[]string,
	error,
) {
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	dstCommittees := []byte{}
	res := []string{}
	switch msgType {
	case topic.CmdBlockBeacon:
		dstCommittees = pubsub.SupportShards
	case topic.CmdBlkShardToBeacon:
		dstCommittees = []byte{common.BEACONID}
	case topic.CmdCrossShard:
		dstCommitteeID := topic.GetCommitteeIDOfTopic(topicReceived)
		dstCommittees = []byte{dstCommitteeID}
	}

	for _, cid := range dstCommittees {
		pubTopic := topic.GetTopicForPub(true, msgType, cid)
		res = append(res, pubTopic)
	}
	return res, nil
}

func (pubsub *PubSubManager) PublishPeerStateToNode() {
	listStateData := [][]byte{}
	for cID, committeeState := range pubsub.BlockChainData.ListMsgPeerStateOfShard {
		pubTopic := topic.GetTopicForPub(true, topic.CmdPeerState, cID)
		for _, stateData := range committeeState {
			listStateData = append(listStateData, stateData)
		}
		err := PublishDataWithTopic(pubsub.FloodMachine, []string{pubTopic}, listStateData, OneTopicNData)
		if err != nil {
			logger.Errorf("Publish Peer state to Committee %v return error %v", cID, err)
		}
	}
}

// func (pubsub *PubSubManager) SubKnownTopics() error {
// 	for _, cID := range pubsub.SupportShards {
// 		topicPub := topic.GetTopicForPub(true, )
// 	}
// 	return nil
// }
