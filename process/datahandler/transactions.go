package datahandler

import (
	"highway/common"
	"highway/process/topic"
	"time"

	cCommon "github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
)

type TxHandler struct {
	FromNode bool
	PubSub   *libp2p.PubSub
	// Locker   *sync.Mutex
}

const MaxDuration = 4 * time.Hour

func (handler *TxHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	var topicPubs []string
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	cID := topic.GetCommitteeIDOfTopic(topicReceived)
	locktime, err := common.GetTransactionTimestamp(dataReceived.Data)
	if err == nil {
		dur := time.Now().Unix() - locktime
		if dur > int64(MaxDuration.Seconds()) {
			return errors.Errorf("Tx from msg %v too old, time stamp %v", cCommon.HashH(dataReceived.Data).String(), locktime)
		}
	} else {
		logger.Error(err)
	}
	if handler.FromNode {
		logger.Infof("[msgtx] received tx from topic %v, data received %v", topicReceived, cCommon.HashH(dataReceived.Data).String())
		topicPub := topic.Handler.GetHWPubSubOutSideFromMsg(msgType, cID)
		topicPubs = append(topicPubs, topicPub)
	} else {
		topicPubs = topic.Handler.GetHWPubTopicsFromMsg(msgType, cID)
	}
	logger.Debugf("[msgtx] Handle topic %v, isInside %v:", topicReceived, handler.FromNode)
	for _, topicPub := range topicPubs {
		logger.Debugf("[msgtx]\tPublish topic %v", topicPub)
		err := handler.PubSub.Publish(topicPub, dataReceived.GetData())
		if err != nil {
			logger.Errorf("Publish topic %v return error: %v", topicPub, err)
		}
	}
	return nil
}
