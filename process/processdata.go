package process

import (
	fmt "fmt"
	"highway/common"
	logger "highway/customizelog"
	"highway/process/topic"

	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

func PeriodicalPublish(
	pubMachine *p2pPubSub.PubSub,
	// chainData *ChainData,
	messageType string,
	committeeID byte,
	data []byte,
) error {
	pubTopic := topic.GetTopic4ProxyPub(committeeID, messageType)
	logger.Infof("Publish Topic %v", pubTopic)
	err := pubMachine.Publish(pubTopic, data)
	if err != nil {
		fmt.Printf("Publish peerstate to committeeID %v return error :%v\n", committeeID, err)
	}
	return err //handle error later
}

func ProcessNPublishDataFromTopic(
	pubMachine *p2pPubSub.PubSub,
	topicReceived string,
	data []byte,
	supportCommittee []byte,
) error {
	// logger.Info("ProcessNPublishDataFromTopic")
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	logger.Info(topicReceived, msgType)
	switch msgType {
	case topic.CmdBlockBeacon: //, topic.CmdBlockShard, topic.CmdBlkShardToBeacon: //, topic.CmdAddr:
		logger.Infof("Process and publish blockbeacon to shards %v", supportCommittee)
		for _, committeeID := range supportCommittee {
			if committeeID != common.BEACONID {
				go PeriodicalPublish(pubMachine, topic.CmdBlockBeacon, committeeID, data)
			}
		}
	case /*topic.CmdBlockShard, */ topic.CmdBlkShardToBeacon:
		logger.Infof("Process and publish shard to beacon")
		go PeriodicalPublish(pubMachine, topic.CmdBlkShardToBeacon, common.BEACONID, data)
	case topic.CmdCrossShard:
		dstCommitteeID := topic.GetCommitteeIDOfTopic(topicReceived)
		logger.Infof("Process and publish crossshard to shard %v", dstCommitteeID)
		go PeriodicalPublish(pubMachine, topic.CmdCrossShard, dstCommitteeID, data)
	}
	return nil
}

func testParsePeerState(data []byte) {

}
