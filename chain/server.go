package chain

import (
	"context"
	"errors"
	"highway/common"
	"highway/process"
	"highway/process/topic"
	"highway/proto"

	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/grpc"
)

func (s *Server) Register(
	ctx context.Context,
	req *proto.RegisterRequest,
) (
	*proto.RegisterResponse,
	error,
) {
	// TODO(@akk0r0kamui): auth committee pubkey and peerID
	logger.Debugf("Receive new request from %v via gRPC", req.GetPeerID())
	committeeID, err := s.hc.chainData.GetCommitteeIDOfValidator(req.GetCommitteePublicKey())
	if err != nil {
		return nil, err
	}
	pairs, err := s.processListWantedMessageOfPeer(req.GetWantedMessages(), committeeID, req.GetPeerID())
	// logger.Info(pairs)
	if err != nil {
		return nil, err
	}
	//	return &ProxyRegisterResponse{Pair: pairs}, nil

	// Notify HighwayClient of a new peer to request data later if possible
	pid, err := peer.IDB58Decode(req.PeerID)
	s.hc.chainData.UpdatePeerIDOfCommitteePubkey(req.GetCommitteePublicKey(), &pid)
	cid := int(committeeID)

	if err == nil {
		s.m.newPeers <- PeerInfo{ID: pid, CID: cid}
	} else {
		logger.Errorf("Invalid peerID: %v", req.PeerID)
	}

	// Return response to node
	role := process.GetUserRole(cid)
	return &proto.RegisterResponse{Pair: pairs, Role: role}, nil
}

func (s *Server) GetBlockShardByHeight(ctx context.Context, req *proto.GetBlockShardByHeightRequest) (*proto.GetBlockShardByHeightResponse, error) {
	logger.Debugf("Receive GetBlockShardByHeight request, shard %v, from height %v to %v", req.Shard, req.FromHeight, req.ToHeight)
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
	return &proto.GetBlockShardByHeightResponse{Data: data}, nil
}

func (s *Server) GetBlockShardByHash(ctx context.Context, req *proto.GetBlockShardByHashRequest) (*proto.GetBlockShardByHashResponse, error) {
	logger.Debug("Receive GetBlockShardByHash request")
	return nil, nil
}

func (s *Server) GetBlockBeaconByHeight(ctx context.Context, req *proto.GetBlockBeaconByHeightRequest) (*proto.GetBlockBeaconByHeightResponse, error) {
	logger.Debugf("Receive GetBlockBeaconByHeight request, from height %v to %v", req.FromHeight, req.ToHeight)
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
	return &proto.GetBlockBeaconByHeightResponse{Data: data}, nil
}

func (s *Server) GetBlockShardToBeaconByHeight(
	ctx context.Context,
	req *proto.GetBlockShardToBeaconByHeightRequest,
) (
	*proto.GetBlockShardToBeaconByHeightResponse,
	error,
) {
	logger.Debugf("Receive GetBlockShardToBeaconByHeight request, from shard = %v, from height %v to %v", req.FromShard, req.FromHeight, req.ToHeight)
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
	return &proto.GetBlockShardToBeaconByHeightResponse{Data: data}, nil
}

func (s *Server) GetBlockBeaconByHash(ctx context.Context, req *proto.GetBlockBeaconByHashRequest) (*proto.GetBlockBeaconByHashResponse, error) {
	logger.Debug("Receive GetBlockBeaconByHash request")
	return nil, nil
}

func (s *Server) GetBlockCrossShardByHeight(ctx context.Context, req *proto.GetBlockCrossShardByHeightRequest) (*proto.GetBlockCrossShardByHeightResponse, error) {
	logger.Debugf("Receive GetBlockCrossShardByHeight request, from Shard %v to shard %v, from height %v to %v", req.FromShard, req.ToShard, req.FromHeight, req.ToHeight)
	data, err := s.hc.GetBlockCrossShardByHeight(
		req.FromShard,
		req.ToShard,
		req.Specific,
		req.FromHeight,
		req.ToHeight,
		req.Heights,
		req.FromPool,
	)
	if err != nil {
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &proto.GetBlockCrossShardByHeightResponse{Data: data}, nil
}

func (s *Server) GetBlockCrossShardByHash(ctx context.Context, req *proto.GetBlockCrossShardByHashRequest) (*proto.GetBlockCrossShardByHashResponse, error) {
	logger.Debug("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

type Server struct {
	m  *Manager
	hc *Client
}

func RegisterServer(m *Manager, gs *grpc.Server, hc *Client) {
	s := &Server{hc: hc, m: m}
	proto.RegisterHighwayServiceServer(gs, s)
}

func (s *Server) processListWantedMessageOfPeer(
	msgs []string,
	committeeID byte,
	peerID string,
) (
	[]*proto.MessageTopicPair,
	error,
) {
	pairs := []*proto.MessageTopicPair{}
	for _, m := range msgs {
		pair, err := s.generateResponseTopic(&process.GlobalPubsub, committeeID, m)
		if err != nil {
			logger.Infof("generateResponseTopic failed, error: %v", err.Error())
		} else {
			pairs = append(pairs, pair)
		}
	}
	return pairs, nil
}

func (s *Server) generateResponseTopic(
	pubsubManager *process.PubSubManager,
	committeeID byte,
	msg string,
) (
	*proto.MessageTopicPair,
	error,
) {

	// TODO Update this stupid code and the way to check duplicate sub message

	var responseTopic []string
	var actOfTopic []proto.MessageTopicPair_Action
	topicGenerator := new(topic.InsideTopic)
	switch msg {
	case topic.CmdCrossShard:
		// handle error later
		responseTopic = make([]string, common.NumberOfShard+1)
		actOfTopic = make([]proto.MessageTopicPair_Action, common.NumberOfShard+1)
		err := topicGenerator.FromMessageType(int(committeeID), msg)
		if err != nil {
			return nil, err
		}
		// responseTopic[common.NumberOfShard] = topicGenerator.GetTopic4ProxyPub()
		responseTopic[common.NumberOfShard] = topic.GetTopicForSub(false, msg, common.NumberOfShard)
		actOfTopic[common.NumberOfShard] = proto.MessageTopicPair_SUB
		for committeeID := common.NumberOfShard - 1; committeeID >= 0; committeeID-- {
			topicGenerator.CommitteeID = byte(committeeID)
			// logger.Info(committeeID, len(responseTopic))
			topic4HighwaySub := topic.GetTopicForSub(true, msg, byte(committeeID))
			responseTopic[committeeID] = topic4HighwaySub
			if !pubsubManager.HasTopic(topic4HighwaySub) {
				pubsubManager.GRPCMessage <- topic4HighwaySub
			}
			actOfTopic[committeeID] = proto.MessageTopicPair_PUB
		}
	case topic.CmdBFT:
		responseTopic = make([]string, 1)
		actOfTopic = make([]proto.MessageTopicPair_Action, 1)
		err := topicGenerator.FromMessageType(int(committeeID), msg)
		if err != nil {
			return nil, err
		}
		topic4HighwaySub := topic.GetTopicForPubSub(msg, committeeID)
		responseTopic[0] = topic4HighwaySub
		if !pubsubManager.HasTopic(topic4HighwaySub) {
			pubsubManager.GRPCMessage <- topic4HighwaySub
		}
		actOfTopic[0] = proto.MessageTopicPair_PUBSUB
	case topic.CmdPeerState, topic.CmdBlockBeacon, topic.CmdBlkShardToBeacon, topic.CmdBlockShard:
		responseTopic = make([]string, 2)
		actOfTopic = make([]proto.MessageTopicPair_Action, 2)
		err := topicGenerator.FromMessageType(int(committeeID), msg)
		if err != nil {
			return nil, err
		}
		// topic4HighwaySub := topicGenerator.GetTopic4ProxySub()
		topic4HighwaySub := topic.GetTopicForSub(true, msg, byte(committeeID))
		responseTopic[0] = topic4HighwaySub
		actOfTopic[0] = proto.MessageTopicPair_PUB
		if !pubsubManager.HasTopic(topic4HighwaySub) {
			pubsubManager.GRPCMessage <- topic4HighwaySub
		}
		responseTopic[1] = topic.GetTopicForSub(false, msg, byte(committeeID))
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
