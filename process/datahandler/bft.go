package datahandler

import libp2p "github.com/incognitochain/go-libp2p-pubsub"

type BFTHandler struct {
}

func (handler *BFTHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	//Do not need to process cuz inside and outside using the same topic bft-<cid>-
	return nil
}
