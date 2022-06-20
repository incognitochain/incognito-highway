package datahandler

import (
	"highway/process/topic"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
	cCommon "github.com/incognitochain/incognito-chain/common"
)

type BlkShardHandler struct {
	FromNode bool
	PubSub   *libp2p.PubSub
	// Locker   *sync.Mutex
}

func (handler *BlkShardHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	var topicPubs []string
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	cID := topic.GetCommitteeIDOfTopic(topicReceived)
	if handler.FromNode {
		logger.Infof("[msgblockshard] received block from topic %v, data received %v", topicReceived, cCommon.HashH(dataReceived.Data).String())
		topicPub := topic.Handler.GetHWPubSubOutSideFromMsg(msgType, cID)
		topicPubs = append(topicPubs, topicPub)
	} else {
		topicPubs = topic.Handler.GetHWPubTopicsFromMsg(msgType, cID)
		topicPubs = topic.Handler.GetHWPubTopicsFromMsg(msgType, topic.NoCIDInTopic)
	}
	for _, topicPub := range topicPubs {
		logger.Debugf("[msgblockshard]\tPublish topic %v", topicPub)
		err := handler.PubSub.Publish(topicPub, dataReceived.GetData())
		if err != nil {
			logger.Errorf("Publish topic %v return error: %v", topicPub, err)
		}
	}
	return nil
}
