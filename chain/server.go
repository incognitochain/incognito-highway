package chain

import (
	"context"
	"highway/common"
	"highway/process/topic"
	"highway/proto"

	peer "github.com/libp2p/go-libp2p-peer"
	"google.golang.org/grpc"
)

/*
func (s *Server) Register(
	ctx context.Context,
	req *proto.RegisterRequest,
) (
	*proto.RegisterResponse,
	error,
) {
	logger.Debugf("Receive new request from %v via gRPC", req.GetPeerID())
	committeeID, err := s.hc.chainData.GetCommitteeIDOfValidator(req.GetCommitteePublicKey())
	isValidator := true
	if err != nil {
		// return nil, err
		isValidator = false
	}
	pairs, err := s.processListWantedMessageOfPeer(req.GetWantedMessages(), isValidator, committeeID, req.GetPeerID())
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
	return &proto.RegisterResponse{Pair: pairs}, nil
}
*/

func (s *Server) Register(
	ctx context.Context,
	req *proto.RegisterRequest,
) (
	*proto.RegisterResponse,
	error,
) {
	// TODO Add list of committeeID, which node wanna sub/pub,..., into register request
	role, cID := s.hc.chainData.GetCommitteeInfoOfPublicKey(req.GetCommitteePublicKey())
	pairs, err := s.processListWantedMessageOfPeer(req.GetWantedMessages(), role, cID)
	if err != nil {
		return nil, err
	}

	// Notify HighwayClient of a new peer to request data later if possible
	pid, err := peer.IDB58Decode(req.PeerID)
	s.hc.chainData.UpdatePeerIDOfCommitteePubkey(req.GetCommitteePublicKey(), &pid)

	if err == nil {
		s.m.newPeers <- PeerInfo{ID: pid, CID: int(cID)}
	} else {
		logger.Errorf("Invalid peerID: %v", req.PeerID)
	}
	return &proto.RegisterResponse{Pair: pairs}, nil
}

func (s *Server) GetBlockShardByHeight(ctx context.Context, req *proto.GetBlockShardByHeightRequest) (*proto.GetBlockShardByHeightResponse, error) {
	// logger.Println("Receive GetBlockShardByHeight request")
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
	// logger.Debug("Receive GetBlockBeaconByHeight request")
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
	logger.Debug("Receive GetBlockShardToBeaconByHeight request")
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
	logger.Debug("Receive GetBlockCrossShardByHeight request")
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
	role byte,
	committeeID byte,
) (
	[]*proto.MessageTopicPair,
	error,
) {
	pairs := []*proto.MessageTopicPair{}
	msgAndCID := map[string][]int{}
	for _, m := range msgs {
		cIDs := []int{}
		if role != common.CANDIDATE {
			// TODO Support non-validator, add list wanted shard into register request
			cIDs = []int{0}
		} else {
			cIDs = []int{int(committeeID)}
		}
		msgAndCID[m] = cIDs
	}
	// TODO handle error here
	pairs = topic.Handler.GetListTopicPairForNode(role, msgAndCID)
	return pairs, nil
}
