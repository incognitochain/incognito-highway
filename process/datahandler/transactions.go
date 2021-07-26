package datahandler

import (
	"highway/common"
	"highway/process/topic"
	"time"

	cCommon "github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
)

type TxHandler struct {
	FromNode bool
	PubSub   *libp2p.PubSub
	Cacher   *cache.Cache
}

const MaxDuration = 3 * time.Hour

func (handler *TxHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	var topicPubs []string
	dataHash := cCommon.HashH(dataReceived.Data).String()
	if _, isExist := handler.Cacher.Get(dataHash); isExist {
		return errors.Errorf("Received old data, dataHash %v", dataHash)
	}
	handler.Cacher.Set(dataHash, nil, 3*time.Hour)
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	cID := topic.GetCommitteeIDOfTopic(topicReceived)
	msg, err := common.ParseMsgChainData(dataReceived.Data)
	if err == nil {
		var tx metadata.Transaction
		if msgTx, ok := msg.(*wire.MessageTx); ok {
			tx = msgTx.Transaction
		} else {
			if msgTx, ok := msg.(*wire.MessageTxPrivacyToken); ok {
				tx = msgTx.Transaction
			} else {
				logger.Errorf("Can not parse transaction")
				tx = nil
			}
		}
		if tx != nil {
			dur := time.Now().Unix() - tx.GetLockTime()
			if dur > int64(MaxDuration.Seconds()) {
				return errors.Errorf("Tx %v too old, time stamp %v", tx.Hash().String(), tx.GetLockTime())
			}
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
