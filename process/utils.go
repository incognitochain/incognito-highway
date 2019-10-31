package process

import (
	"errors"
	"highway/common"
	"highway/process/topic"
	"highway/proto"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
)

var CommitteePubkeyByPeerID map[peer.ID]string
var updatePKbyPIDLocker sync.Mutex

func UpdatePeerIDOfCommitteePubkey(
	candidate string,
	peerID *peer.ID,
) {
	updatePKbyPIDLocker.Lock()
	CommitteePubkeyByPeerID[*peerID] = candidate
	updatePKbyPIDLocker.Unlock()
}

func GenerateResponseTopic(pubsubManager *PubSubManager, nodePK, msg string) (*proto.MessageTopicPair, error) {
	var responseTopic []string
	var actOfTopic []proto.MessageTopicPair_Action
	topicGenerator := new(topic.InsideTopic)
	switch msg {
	case topic.CmdCrossShard:
		// handle error later
		responseTopic = make([]string, common.NumberOfShard+1)
		actOfTopic = make([]proto.MessageTopicPair_Action, common.NumberOfShard+1)
		err := topicGenerator.FromMessageType(nodePK, msg)
		if err != nil {
			return nil, err
		}
		responseTopic[common.NumberOfShard] = topicGenerator.GetTopic4ProxyPub()
		actOfTopic[common.NumberOfShard] = proto.MessageTopicPair_SUB
		for committeeID := common.NumberOfShard - 1; committeeID >= 0; committeeID-- {
			topicGenerator.CommitteeID = byte(committeeID)
			// fmt.Println(committeeID, len(responseTopic))
			responseTopic[committeeID] = topicGenerator.GetTopic4ProxySub()
			pubsubManager.GRPCMessage <- responseTopic[committeeID]
			actOfTopic[committeeID] = proto.MessageTopicPair_PUB
		}
	case topic.CmdBFT:
		responseTopic = make([]string, 1)
		actOfTopic = make([]proto.MessageTopicPair_Action, 1)
		err := topicGenerator.FromMessageType(nodePK, msg)
		if err != nil {
			return nil, err
		}
		responseTopic[0] = topicGenerator.ToString()
		pubsubManager.GRPCMessage <- responseTopic[0]
		actOfTopic[0] = proto.MessageTopicPair_PUBSUB
	case topic.CmdPeerState, topic.CmdBlockBeacon, topic.CmdBlkShardToBeacon, topic.CmdBlockShard:
		responseTopic = make([]string, 2)
		actOfTopic = make([]proto.MessageTopicPair_Action, 2)
		err := topicGenerator.FromMessageType(nodePK, msg)
		if err != nil {
			return nil, err
		}
		responseTopic[0] = topicGenerator.GetTopic4ProxySub()
		actOfTopic[0] = proto.MessageTopicPair_PUB
		pubsubManager.GRPCMessage <- topicGenerator.GetTopic4ProxySub()
		responseTopic[1] = topicGenerator.GetTopic4ProxyPub()
		actOfTopic[1] = proto.MessageTopicPair_SUB

	default:
		return nil, errors.New("Unknown message type: " + msg)
	}
	return &proto.MessageTopicPair{
		Message: msg,
		Topic:   responseTopic,
		Act:     actOfTopic,
	}, nil
}
