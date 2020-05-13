package datahandler

import (
	"highway/common"
	"highway/process/simulateutils"
	"highway/process/topic"

	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-pubsub"
)

type BFTHandler struct {
	CommitteeInfo   *simulateutils.CommitteeTable
	Scenario        *simulateutils.Scenario
	currentHeight   uint64
	currentPubGroup map[string][]string
	PubSub          *libp2p.PubSub
}

func (handler *BFTHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	//Do not need to process cuz inside and outside using the same topic bft-<cid>-
	srcPKey := topic.GetPubkeyInTopic(topicReceived)
	msg, err := common.ParseMsgChainData(dataReceived.Data)
	if err != nil {
		return err
	}
	msgBFT, ok := msg.(*wire.MessageBFT)
	if !ok {
		return nil
	}
	pubGroup := handler.Scenario.GetPubGroup(msgBFT.ChainKey, uint64(msgBFT.Timestamp))
	if msgBFT.ChainKey == "beacon" {
		logger.Infof("[monitor] srckey %v Block %v Get PubGroup %v", srcPKey, msgBFT.Timestamp, pubGroup)
	}
	if pubGroup != nil {
		for _, dstPKey := range pubGroup[srcPKey] {
			if msgBFT.ChainKey == "beacon" {
				logger.Infof("[monitor] from %v to %v", srcPKey, dstPKey)
			}
			newTopic := topic.SwitchTopicPrivate(topicReceived, dstPKey)
			handler.PubSub.Publish(newTopic, dataReceived.Data)
		}
	} else {
		for _, dstPKey := range handler.CommitteeInfo.GetKeysByKey(srcPKey) {
			newTopic := topic.SwitchTopicPrivate(topicReceived, dstPKey)
			handler.PubSub.Publish(newTopic, dataReceived.Data)
		}
	}
	return nil
}
