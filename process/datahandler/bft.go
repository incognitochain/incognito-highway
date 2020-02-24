package datahandler

import (
	"highway/common"

	libp2p "github.com/libp2p/go-libp2p-pubsub"
)

type BFTHandler struct {
}

func (handler *BFTHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	msgbft, err := common.ParseBFTData(string(dataReceived.Data))
	logger.Infof("[msgbft] Received msg bft, len %v", len(dataReceived.Data))
	if err != nil {
		logger.Infof("[msgbft] Can not parse message bft, error %v", err)
	} else {
		logger.Infof("[msgbft] received Type: %v, ChainKey %v", msgbft.Type, msgbft.ChainKey)
	}
	//Do not need to process cuz inside and outside using the same topic bft-<cid>-
	return nil
}
