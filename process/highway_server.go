package process

import (
	"context"
	"errors"
	fmt "fmt"
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process/topic"

	"github.com/libp2p/go-libp2p-core/peer"
)

func (s *HighwayServer) Register(
	ctx context.Context,
	req *RegisterRequest,
) (
	*RegisterResponse,
	error,
) {
	logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	committeeID, err := s.hc.chainData.GetCommitteeIDOfValidator(req.GetCommitteePublicKey())
	if err != nil {
		return nil, err
	}
	pairs, err := s.processListWantedMessageOfPeer(req.GetWantedMessages(), committeeID, req.GetPeerID())
	logger.Info(pairs)
	if err != nil {
		return nil, err
	}
	// Notify HighwayClient of a new peer to request data later if possible
	// TODO hy remove duplicate peer info object
	pid, err := peer.IDB58Decode(req.PeerID)
	s.hc.chainData.UpdatePeerIDOfCommitteePubkey(req.GetCommitteePublicKey(), &pid)
	cid := int(committeeID)
	if err == nil {
		s.hc.NewPeers <- PeerInfo{ID: pid, CID: cid}
	} else {
		logger.Errorf("Invalid peerID: %v", req.PeerID)
	}
	return &RegisterResponse{Pair: pairs}, nil
}

func (s *HighwayServer) GetBlockShardByHeight(ctx context.Context, req *GetBlockShardByHeightRequest) (*GetBlockShardByHeightResponse, error) {
	logger.Println("Receive GetBlockShardByHeight request")
	// TODO(@0xbunyip): check if block in cache

	// Call node to get blocks
	// TODO(@0xbunyip): use fromPool
	data, err := s.hc.GetBlockShardByHeight(
		req.Shard,
		req.Specific,
		req.FromHeight,
		req.ToHeight,
		req.Heights,
	)
	if err != nil {
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &GetBlockShardByHeightResponse{Data: data}, nil
}

func (s *HighwayServer) GetBlockShardByHash(ctx context.Context, req *GetBlockShardByHashRequest) (*GetBlockShardByHashResponse, error) {
	logger.Println("Receive GetBlockShardByHash request")
	return nil, nil
}

func (s *HighwayServer) GetBlockBeaconByHeight(ctx context.Context, req *GetBlockBeaconByHeightRequest) (*GetBlockBeaconByHeightResponse, error) {
	logger.Println("Receive GetBlockBeaconByHeight request")
	// TODO(@0xbunyip): check if block in cache

	// Call node to get blocks
	// TODO(@0xbunyip): use fromPool
	data, err := s.hc.GetBlockBeaconByHeight(
		req.Specific,
		req.FromHeight,
		req.ToHeight,
		req.Heights,
	)
	if err != nil {
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &GetBlockBeaconByHeightResponse{Data: data}, nil
}

func (s *HighwayServer) GetBlockBeaconByHash(ctx context.Context, req *GetBlockBeaconByHashRequest) (*GetBlockBeaconByHashResponse, error) {
	logger.Println("Receive GetBlockBeaconByHash request")
	return nil, nil
}

func (s *HighwayServer) GetBlockCrossShardByHeight(ctx context.Context, req *GetBlockCrossShardByHeightRequest) (*GetBlockCrossShardByHeightResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHeight request")
	data, err := s.hc.GetBlockShardByHeight(
		req.GetFromShard(),
		req.Specific,
		req.FromHeight,
		req.ToHeight,
		req.Heights,
	)
	if err != nil {
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &GetBlockCrossShardByHeightResponse{Data: data}, nil
	return nil, nil
}

func (s *HighwayServer) GetBlockShardToBeaconByHeight(
	ctx context.Context,
	req *GetBlockShardToBeaconByHeightRequest,
) (
	*GetBlockShardToBeaconByHeightResponse,
	error,
) {
	logger.Println("Receive GetBlockCrossShardByHeight request")
	data, err := s.hc.GetBlockShardToBeaconByHeight(
		req.GetFromShard(),
		req.Specific,
		req.FromHeight,
		req.ToHeight,
		req.Heights,
	)
	if err != nil {
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &GetBlockShardToBeaconByHeightResponse{Data: data}, nil
	return nil, nil
}

func (s *HighwayServer) GetBlockCrossShardByHash(ctx context.Context, req *GetBlockCrossShardByHashRequest) (*GetBlockCrossShardByHashResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

type HighwayServer struct {
	hc *HighwayClient
}

func RunHighwayServer(h *p2p.Host, chainData *ChainData, hc *HighwayClient) {
	s := &HighwayServer{
		hc: hc,
	}
	RegisterHighwayServiceServer(h.GRPC.GetGRPCServer(), s)
	go h.GRPC.Serve() // NOTE: must serve after registering all services
}

func (s *HighwayServer) processListWantedMessageOfPeer(
	msgs []string,
	committeeID byte,
	peerID string,
) (
	[]*MessageTopicPair,
	error,
) {
	pairs := []*MessageTopicPair{}
	for _, m := range msgs {
		pair, err := s.generateResponseTopic(&GlobalPubsub, committeeID, m)
		if err != nil {
			logger.Infof("generateResponseTopic failed, error: %v", err.Error())
			pair = &MessageTopicPair{
				Message: m,
				Topic:   []string{},
				Act:     []MessageTopicPair_Action{},
			}
			return nil, err
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func (s *HighwayServer) generateResponseTopic(
	pubsubManager *PubSubManager,
	committeeID byte,
	msg string,
) (
	*MessageTopicPair,
	error,
) {
	var responseTopic []string
	var actOfTopic []MessageTopicPair_Action
	topicGenerator := new(topic.InsideTopic)
	switch msg {
	case topic.CmdCrossShard:
		// handle error later
		responseTopic = make([]string, common.NumberOfShard+1)
		actOfTopic = make([]MessageTopicPair_Action, common.NumberOfShard+1)
		err := topicGenerator.FromMessageType(int(committeeID), msg)
		if err != nil {
			return nil, err
		}
		responseTopic[common.NumberOfShard] = topicGenerator.GetTopic4ProxyPub()
		actOfTopic[common.NumberOfShard] = MessageTopicPair_SUB
		for committeeID := common.NumberOfShard - 1; committeeID >= 0; committeeID-- {
			topicGenerator.CommitteeID = byte(committeeID)
			fmt.Println(committeeID, len(responseTopic))
			responseTopic[committeeID] = topicGenerator.GetTopic4ProxySub()
			pubsubManager.GRPCMessage <- responseTopic[committeeID]
			actOfTopic[committeeID] = MessageTopicPair_PUB
		}
	case topic.CmdBFT:
		responseTopic = make([]string, 1)
		actOfTopic = make([]MessageTopicPair_Action, 1)
		err := topicGenerator.FromMessageType(int(committeeID), msg)
		if err != nil {
			return nil, err
		}
		responseTopic[0] = topicGenerator.ToString()
		pubsubManager.GRPCMessage <- responseTopic[0]
		actOfTopic[0] = MessageTopicPair_PUBSUB
	case topic.CmdPeerState, topic.CmdBlockBeacon, topic.CmdBlkShardToBeacon, topic.CmdBlockShard:
		responseTopic = make([]string, 2)
		actOfTopic = make([]MessageTopicPair_Action, 2)
		err := topicGenerator.FromMessageType(int(committeeID), msg)
		if err != nil {
			return nil, err
		}
		responseTopic[0] = topicGenerator.GetTopic4ProxySub()
		actOfTopic[0] = MessageTopicPair_PUB
		pubsubManager.GRPCMessage <- topicGenerator.GetTopic4ProxySub()
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
