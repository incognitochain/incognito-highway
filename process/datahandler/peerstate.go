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
	logger.Debugf("[msgpeerstate] Handle topic %v, isInside %v:", topicReceived, handler.FromNode)
	if handler.FromNode {
		logger.Debugf("[msgpeerstate] Received Peerstate from %v", dataReceived.GetFrom())
		topicPub := topic.Handler.GetHWPubSubOutSideFromMsg(msgType, cID)
		topicPubs = append(topicPubs, topicPub)
		for _, topicPub := range topicPubs {
			err := handler.PubSub.Publish(topicPub, dataReceived.GetData())
			if err != nil {
				logger.Errorf("Publish topic %v return error: %v", topicPub, err)
			}
		}
	} else {
		logger.Debugf("[msgpeerstate] Received Peerstate from highway %v", dataReceived.GetFrom())
		err := handler.BlockchainData.UpdatePeerStateFromHW(dataReceived.GetFrom().String(), dataReceived.GetData())
		if err != nil {
			logger.Errorf("Update peerState when received message from %v return error %v ", dataReceived.GetFrom(), dataReceived.GetData())
		}
	}
	return nil
}
