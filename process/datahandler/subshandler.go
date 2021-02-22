package datahandler

import (
	"context"
	"fmt"
	"highway/chaindata"
	"highway/common"
	"highway/process/topic"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/patrickmn/go-cache"
)

type SubsHandler struct {
	PubSub         *libp2p.PubSub
	BlockchainData *chaindata.ChainData
	dataHandler    DataHandler
	Cacher         *cache.Cache
	FromInside     bool
}

func (handler *SubsHandler) HandlerNewSubs(subs *libp2p.Subscription) error {
	var err error
	if handler.dataHandler == nil {
		handler.dataHandler, err = handler.GetDataHandler(subs.Topic(), handler.FromInside)
		if err != nil {
			return err
		}
	}
	for {
		data, err := subs.Next(context.Background())
		if (err == nil) && (data != nil) {
			dataBytes := data.GetData()
			data4cache := common.NewKeyForCacheDataOfTopic(subs.Topic(), dataBytes)
			if _, isExist := handler.Cacher.Get(string(data4cache)); isExist {
				continue
			}
			handler.Cacher.Set(string(data4cache), nil, common.MaxTimeKeepPubSubData)
			go func() {
				err := handler.dataHandler.HandleDataFromTopic(subs.Topic(), *data)
				if err != nil {
					logger.Errorf("Can not process data from topic %v, handler return error %v", subs.Topic(), err)
				}
			}()
		} else {
			if err == nil {
				err = fmt.Errorf("Received nil data form topic %v", subs.Topic())
			}
			logger.Error(err)
		}
	}
}

func (handler *SubsHandler) GetDataHandler(
	topicReceived string,
	forInside bool,
) (
	DataHandler,
	error,
) {
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	switch msgType {
	case topic.CmdBlockBeacon:
		return &BlkBeaconHandler{
			FromNode: forInside,
			PubSub:   handler.PubSub,
		}, nil
	case topic.CmdBlockShard:
		return &BlkShardHandler{
			FromNode: forInside,
			PubSub:   handler.PubSub,
		}, nil
	case topic.CmdBlkShardToBeacon, topic.CmdFinishSync, topic.CmdCrossShard:
		return &BlkCrossCommitteeHandler{
			FromNode: forInside,
			PubSub:   handler.PubSub,
		}, nil
	case topic.CmdTx, topic.CmdCustomToken, topic.CmdPrivacyCustomToken:
		return &TxHandler{
			FromNode: forInside,
			PubSub:   handler.PubSub,
		}, nil
	case topic.CmdPeerState:
		return &PeerStateHandler{
			FromNode:       forInside,
			PubSub:         handler.PubSub,
			BlockchainData: handler.BlockchainData,
		}, nil
	case topic.CmdBFT:
		return &BFTHandler{}, nil
	default:
		return nil, fmt.Errorf("Handler for msg %v can not found", msgType)
	}
}
