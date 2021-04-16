package datahandler

import (
	"fmt"
	"highway/common"
	"highway/process/simulateutils"
	"highway/process/topic"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/incognitochain/incognito-chain/consensus/blsbftv2"
	"github.com/incognitochain/incognito-chain/wire"
)

type BFTHandler struct {
	CommitteeInfo   *simulateutils.CommitteeTable
	Scenario        *simulateutils.Scenario
	currentHeight   uint64
	currentPubGroup map[string][]string
	PubSub          *libp2p.PubSub
	FMaker          simulateutils.ForkMaker
}

func (handler *BFTHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	//Do not need to process cuz inside and outside using the same topic bft-<cid>-
	cID := topic.GetCommitteeIDOfTopic(topicReceived)
	msg, err := common.ParseMsgChainData(dataReceived.Data)
	if err != nil {
		return err
	}
	fmt.Printf("f.PubSub %v\n", &handler.PubSub)
	msgBFT, ok := msg.(*wire.MessageBFT)
	logger.Infof("[debugreceivedbft] Received bft msg type %v from cID %v", msgBFT.Type, cID)
	srcPKey := msgBFT.PeerID
	fmt.Println(srcPKey)
	if !ok {
		return nil
	}

	if handler.FMaker != nil {
		logger.Infof("debuglock Check trigger cID %v", cID)
		if fork := handler.FMaker.IsTrigger(msg, int(cID)); fork {
			logger.Infof("debuglock Check trigger cID %v done", cID)
			switch msgBFT.Type {
			case blsbftv2.MSG_PROPOSE:
				logger.Infof("[debugfork] cID %v Received propose from key %v %v %v %v", cID, srcPKey, msgBFT.ChainKey, msgBFT.TimeSlot, msgBFT.Timestamp)
				handler.FMaker.PBFTInbox(cID) <- simulateutils.MsgBFT{
					ProposeIndex: handler.CommitteeInfo.PubKeyIdx[srcPKey],
					Msg:          msg,
					Topic:        topicReceived,
				}
				logger.Infof("[debugfork] cID %v Sent propose from key %v ", cID, srcPKey)
			case blsbftv2.MSG_VOTE:
				msgV := simulateutils.ParseBFTVote(msgBFT)
				logger.Infof("[debugfork] %v Received vote from %v, block %v\n", cID, srcPKey, msgV.BlkHash)
				handler.FMaker.VBFTInbox(cID) <- simulateutils.MsgBFT{
					ProposeIndex: handler.CommitteeInfo.PubKeyIdx[srcPKey],
					Msg:          msg,
					Topic:        topicReceived,
				}
				logger.Infof("[debugfork] %v Sent vote from %v, block %v\n", cID, srcPKey, msgV.BlkHash)
			}
			return nil
		}
		logger.Infof("debuglock Check trigger cID %v done", cID)
	}

	logger.Infof("[fork] cID %v Fork maker is nil or not trigger fork\n", cID)
	listPK := handler.CommitteeInfo.PubKeyBySID[byte(cID)]
	logger.Infof("[monitor] srckey %v Block %v CID %v Get list pubkey %v", srcPKey, cID, msgBFT.TimeSlot, listPK)
	// fmt.Println(handler.CommitteeInfo.SIDByPubKey)
	// fmt.Println(listPK)
	for _, dstPKey := range listPK {
		newTopic := topic.SwitchTopicPrivate(topicReceived, dstPKey)
		err := handler.PubSub.Publish(newTopic, dataReceived.Data)
		if err != nil {
			logger.Errorf("Publish message to topic %v return err %v", newTopic, err)
			return err
		}
	}
	return nil
}
