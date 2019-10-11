package process

import (
	"context"
	"errors"
	fmt "fmt"
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process/topic"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
)

// var CommitteeIDByPeerID map[string]peer.ID
var CommitteePubkeyByPeerID map[peer.ID]string
var IsPublicIP map[peer.ID]bool
var updatePKbyPIDLocker sync.Mutex

func ProcessConnection(h *p2p.Host) {
	s := &GRPCService_Server{}
	RegisterProxyRegisterServiceServer(h.GRPC.GetGRPCServer(), s)
}

func UpdatePeerIDOfCommitteePubkey(
	candidate string,
	peerID *peer.ID,
) {
	updatePKbyPIDLocker.Lock()
	CommitteePubkeyByPeerID[*peerID] = candidate
	updatePKbyPIDLocker.Unlock()
}

func CheckPublicIPOfPeer(peerID peer.ID) (bool, error) {
	return true, nil
}

func (s *GRPCService_Server) ProxyRegister(
	ctx context.Context,
	req *ProxyRegisterMsg,
) (
	*ProxyRegisterResponse,
	error,
) {
	logger.Infof("Receive new request from gRPC")
	pairs := []*MessageTopicPair{}
	var pair *MessageTopicPair
	var err error
	for _, m := range req.WantedMessages {
		pair, err = generateResponseTopic(&GlobalPubsub, req.CommitteePublicKey, m)
		if err != nil {
			logger.Infof("generateResponseTopic failed, error: %v", err.Error())
			pair = &MessageTopicPair{
				Message: m,
				Topic:   []string{},
				Act:     []MessageTopicPair_Action{},
			}
		}
		peerid, _ := peer.IDB58Decode(req.GetPeerID())
		UpdatePeerIDOfCommitteePubkey(req.GetCommitteePublicKey(), &peerid)

		// if isPublic, exist := IsPublicIP[req.PeerID]; !exist {
		// 	res, err := CheckPublicIPOfPeer(req.PeerID)
		// 	if err != nil {
		// 		logger.Info("Please handle error here")
		// 	} else {
		// 		IsPublicIP[req.PeerID] = res
		// 	}
		// } else {
		// 	if isPublic != IsPublicIP[req.PeerID] {
		// 		IsPublicIP[req.PeerID] = isPublic
		// 	}
		// }
		pairs = append(pairs, pair)
	}
	logger.Info(pairs)
	return &ProxyRegisterResponse{Pair: pairs}, nil
}

type GRPCService_Server struct{}

func generateResponseTopic(pubsubManager *PubSubManager, nodePK, msg string) (*MessageTopicPair, error) {
	var responseTopic []string
	var actOfTopic []MessageTopicPair_Action
	topicGenerator := new(topic.InsideTopic)
	switch msg {
	case topic.CmdCrossShard:
		// handle error later
		responseTopic = make([]string, common.NumberOfShard+1)
		actOfTopic = make([]MessageTopicPair_Action, common.NumberOfShard+1)
		err := topicGenerator.FromMessageType(nodePK, msg)
		if err != nil {
			return nil, err
		}
		responseTopic[common.NumberOfShard] = topicGenerator.GetTopic4ProxyPub()
		actOfTopic[common.NumberOfShard] = MessageTopicPair_SUB
		for committeeID := common.NumberOfShard - 1; committeeID >= 0; committeeID-- {
			topicGenerator.CommitteeID = byte(committeeID)
			fmt.Println(committeeID, len(responseTopic))
			responseTopic[committeeID] = topicGenerator.GetTopic4ProxySub()
			pubsubManager.RegisterMessage <- responseTopic[committeeID]
			actOfTopic[committeeID] = MessageTopicPair_PUB
		}
	case topic.CmdBFT:
		responseTopic = make([]string, 1)
		actOfTopic = make([]MessageTopicPair_Action, 1)
		err := topicGenerator.FromMessageType(nodePK, msg)
		if err != nil {
			return nil, err
		}
		responseTopic[0] = topicGenerator.ToString()
		pubsubManager.RegisterMessage <- responseTopic[0]
		actOfTopic[0] = MessageTopicPair_PUBSUB
	case topic.CmdPeerState, topic.CmdBlockBeacon, topic.CmdBlkShardToBeacon, topic.CmdBlockShard:
		responseTopic = make([]string, 2)
		actOfTopic = make([]MessageTopicPair_Action, 2)
		err := topicGenerator.FromMessageType(nodePK, msg)
		if err != nil {
			return nil, err
		}
		responseTopic[0] = topicGenerator.GetTopic4ProxySub()
		actOfTopic[0] = MessageTopicPair_PUB
		pubsubManager.RegisterMessage <- topicGenerator.GetTopic4ProxySub()
		responseTopic[1] = topicGenerator.GetTopic4ProxyPub()
		actOfTopic[1] = MessageTopicPair_SUB

	default:
		return nil, errors.New("Unknown message type: " + msg)
	}
	return &MessageTopicPair{
		Message: msg,
		Topic:   responseTopic,
		Act:     actOfTopic,
	}, nil
}
