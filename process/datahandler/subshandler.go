package datahandler

import (
	"context"
	"fmt"
	"highway/chaindata"
	"highway/common"
	"highway/database"
	"highway/process/topic"
	"sync"

	libp2p "github.com/libp2p/go-libp2p-pubsub"
)

type SubsHandler struct {
	Locker         *sync.RWMutex
	PubSub         *libp2p.PubSub
	BlockchainData *chaindata.ChainData
	DataHandler    DataHandler
	FromInside     bool
}

func (handler *SubsHandler) HandlerNewSubs(subs *libp2p.Subscription) error {
	var err error
	if handler.DataHandler == nil {
		handler.DataHandler, err = handler.GetDataHandler(subs.Topic(), handler.FromInside)
		if err != nil {
			return err
		}
	}
	for {
		data, err := subs.Next(context.Background())
		if (err == nil) && (data != nil) {
			dataBytes := data.GetData()
			handler.Locker.Lock()
			data4cache := common.NewKeyForCacheDataOfTopic(subs.Topic(), dataBytes)
			if database.IsMarkedData(data4cache) {
				handler.Locker.Unlock()
				continue
			}
			database.MarkData(data4cache)
			handler.Locker.Unlock()
			go func() {
				err := handler.DataHandler.HandleDataFromTopic(subs.Topic(), *data)
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
	case topic.CmdBlkShardToBeacon, topic.CmdCrossShard:
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
