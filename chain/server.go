package chain

import (
	"context"
	"highway/common"
	"highway/process"
	"highway/process/topic"
	"highway/proto"

	"github.com/pkg/errors"

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
	// TODO(@akk0r0kamui): auth committee pubkey and peerID
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

	// Return response to node
	role := process.GetUserRole(cid)
	return &proto.RegisterResponse{Pair: pairs, Role: role}, nil
}
*/

func (s *Server) Register(
	ctx context.Context,
	req *proto.RegisterRequest,
) (
	*proto.RegisterResponse,
	error,
) {
	logger.Infof("Receive Register request, CID %v, peerID %v", req.CommitteeID, req.PeerID)

	// Monitor status
	defer s.reporter.watchRequestCounts("register")

	// TODO Add list of committeeID, which node wanna sub/pub,..., into register request
	role, cID := s.hc.chainData.GetCommitteeInfoOfPublicKey(req.GetCommitteePublicKey())
	cIDs := []int{}
	if role == common.NORMAL {
		reqCIDs := req.GetCommitteeID()
		for _, cid := range reqCIDs {
			cIDs = append(cIDs, int(cid))
		}
	} else {
		cIDs = append(cIDs, cID)
	}
	// logger.Errorf("Received register from -%v- role -%v- cIDs -%v-", req.GetCommitteePublicKey(), role, cIDs)
	pairs, err := s.processListWantedMessageOfPeer(req.GetWantedMessages(), role, cIDs)
	if err != nil {
		logger.Warnf("Couldn't process wantedMsgs: %+v %+v %+v", req.GetWantedMessages(), role, cIDs)
		return nil, err
	}

	r := process.GetUserRole(cID)
	pid, err := peer.IDB58Decode(req.PeerID)
	if err != nil {
		logger.Warnf("Invalid peerID: %v", req.PeerID)
		return nil, err
	}
	pinfo := PeerInfo{ID: pid, Pubkey: req.GetCommitteePublicKey()}
	if role == common.COMMITTEE {
		s.hc.chainData.UpdatePeerIDOfCommitteePubkey(req.GetCommitteePublicKey(), &pid)

		pinfo.CID = int(cID)
		pinfo.Role = r.Role
	} else {
		// TODO(@0xbunyip): support fullnode here (multiple cIDs)
		pinfo.CID = int(cIDs[0])
		pinfo.Role = "normal"
	}
	// Notify HighwayClient of a new peer to request data later if possible
	s.m.newPeers <- pinfo

	// Return response to node
	return &proto.RegisterResponse{Pair: pairs, Role: r}, nil
}

func (s *Server) GetBlockShardByHeight(ctx context.Context, req *proto.GetBlockShardByHeightRequest) (*proto.GetBlockShardByHeightResponse, error) {
	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_shard")

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
	logger.Errorf("Receive GetBlockShardByHash request: %v %x", req.Shard, req.Hashes)
	return nil, errors.New("not supported")
}

func (s *Server) GetBlockBeaconByHeight(ctx context.Context, req *proto.GetBlockBeaconByHeightRequest) (*proto.GetBlockBeaconByHeightResponse, error) {
	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_beacon")

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
	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_shard_to_beacon")

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
	logger.Errorf("Receive GetBlockBeaconByHash request: %x", req.Hashes)
	return nil, errors.New("not supported")
}

func (s *Server) GetBlockCrossShardByHeight(ctx context.Context, req *proto.GetBlockCrossShardByHeightRequest) (*proto.GetBlockCrossShardByHeightResponse, error) {
	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_cross_shard")

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
	logger.Errorf("Receive GetBlockCrossShardByHash request: %d %d %x", req.FromShard, req.ToShard, req.Hashes)
	return nil, errors.New("not supported")
}

type Server struct {
	m  *Manager
	hc *Client

	reporter *Reporter
}

func RegisterServer(m *Manager, gs *grpc.Server, hc *Client, reporter *Reporter) {
	s := &Server{hc: hc, m: m, reporter: reporter}
	proto.RegisterHighwayServiceServer(gs, s)
}

func (s *Server) processListWantedMessageOfPeer(
	msgs []string,
	role byte,
	committeeIDs []int,
) (
	[]*proto.MessageTopicPair,
	error,
) {
	pairs := []*proto.MessageTopicPair{}
	msgAndCID := map[string][]int{}
	for _, m := range msgs {
		msgAndCID[m] = committeeIDs
	}
	// TODO handle error here
	pairs = topic.Handler.GetListTopicPairForNode(role, msgAndCID)
	return pairs, nil
}
