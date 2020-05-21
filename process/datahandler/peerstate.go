package datahandler

import (
	"highway/chaindata"
	"highway/process/topic"

	libp2p "github.com/libp2p/go-libp2p-pubsub"
)

type PeerStateHandler struct {
	FromNode       bool
	PubSub         *libp2p.PubSub
	BlockchainData *chaindata.ChainData
	// Locker         *sync.Mutex
}

func (handler *PeerStateHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	var topicPubs []string
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	cID := topic.GetCommitteeIDOfTopic(topicReceived)
	// logger.Debugf("[msgpeerstate] Handle topic peerstate, cID %v, isInside %v, from %v, from highway: %v", cID, handler.FromNode, dataReceived.GetFrom(), !handler.FromNode)
	if handler.FromNode {
		topicPub := topic.Handler.GetHWPubSubOutSideFromMsg(msgType, cID)
		topicPubs = append(topicPubs, topicPub)
		for _, topicPub := range topicPubs {
			err := handler.PubSub.Publish(topicPub, dataReceived.GetData())
			if err != nil {
				logger.Errorf("Publish topic %v return error: %v", topicPub, err)
			}
		}
	} else {
		err := handler.BlockchainData.UpdatePeerStateFromHW(dataReceived.GetFrom(), dataReceived.GetData(), byte(cID))
		if err != nil {
			logger.Errorf("Update peerState when received message from %v return error %v ", dataReceived.GetFrom(), err)
		}
	}
	return nil
}
