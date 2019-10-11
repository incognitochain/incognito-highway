package process

import (
	"bytes"
	"encoding/json"
	"errors"
	fmt "fmt"
	"highway/common"
	logger "highway/customizelog"
	"highway/process/topic"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

type ChainState struct {
	Timestamp     int64
	Height        uint64
	BlockHash     [32]byte
	BestStateHash [32]byte
}

type CommitteeState map[string][]byte

var AllPeerState map[byte]CommitteeState
var currentAllPeerStateData []byte
var updatePeerLocker sync.Mutex

func initGlobalParams() {
	AllPeerState = map[byte]CommitteeState{}
	for i := byte(0); i < common.NumberOfShard; i++ {
		AllPeerState[i] = map[string][]byte{}
	}
	// listCommittee = append(listCommittee, supportShards...)
	AllPeerState[common.BEACONID] = map[string][]byte{}
	currentAllPeerStateData, _ = json.Marshal(AllPeerState)
	CommitteePubkeyByPeerID = map[peer.ID]string{}
}

func PeriodicalPublish(
	pubMachine *p2pPubSub.PubSub,
	messageType string,
	committeeID byte,
	data []byte,
) error {
	pubTopic := topic.GetTopic4ProxyPub(committeeID, messageType)
	// logger.Infof("Publish Topic %v", pubTopic)
	err := pubMachine.Publish(pubTopic, data)
	if err != nil {
		fmt.Printf("Publish peerstate to committeeID %v return error :%v\n", committeeID, err)
	}
	return nil //handle error later
}

func UpdatePeerState(publisher string, data []byte) error {

	committeeID := common.GetCommitteeIDOfValidator(publisher)
	if (committeeID == -1) || ((committeeID > common.NumberOfShard) && (committeeID < int(common.BEACONID))) {
		logger.Infof("This publisher not belong to current committee -%v- %v %v %v", publisher, common.BEACONID, common.NumberOfShard, committeeID)
		return errors.New("This publisher not belong to current committee")
	}
	updatePeerLocker.Lock()
	if !bytes.Equal(AllPeerState[byte(committeeID)][publisher], data) {
		AllPeerState[byte(committeeID)][publisher] = data
	}
	updatePeerLocker.Unlock()
	return nil
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
		// logger.Infof("Process and publish blockbeacon to shards %v", supportCommittee)
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
