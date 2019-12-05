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

func (s *Server) Register(
	ctx context.Context,
	req *proto.RegisterRequest,
) (
	*proto.RegisterResponse,
	error,
) {
	logger.Infof("Receive Register request, CID %v, peerID %v, role %v", req.CommitteeID, req.PeerID, req.Role)

	// Monitor status
	defer s.reporter.watchRequestCounts("register")

	// TODO Add list of committeeID, which node wanna sub/pub,..., into register request
	reqRole := req.GetRole()
	reqCIDs := req.GetCommitteeID()
	cIDs := []int{}
	for _, cid := range reqCIDs {
		cIDs = append(cIDs, int(cid))
	}

	// Map from user defined role to highway defined role
	role := common.NORMAL // normal node, waiting and pending validators
	if reqRole == common.CommitteeRole {
		role = common.COMMITTEE
	}

	// logger.Errorf("Received register from -%v- role -%v- cIDs -%v-", req.GetCommitteePublicKey(), role, cIDs)
	pairs, err := s.processListWantedMessageOfPeer(req.GetWantedMessages(), role, cIDs)
	if err != nil {
		logger.Warnf("Couldn't process wantedMsgs: %+v %+v %+v", req.GetWantedMessages(), role, cIDs)
		return nil, err
	}

	cID := 0
	if len(cIDs) > 0 {
		cID = cIDs[0] // For validators, cIDs must contain exactly 1 value that is the shard that the they are validating on
	}
	r := process.GetUserRole(reqRole, cID)
	pid, err := peer.IDB58Decode(req.PeerID)
	if err != nil {
		logger.Warnf("Invalid peerID: %v", req.PeerID)
		return nil, err
	}

	key, err := common.PreprocessKey(req.GetCommitteePublicKey())
	if err != nil {
		return nil, err
	}

	pinfo := PeerInfo{ID: pid, Pubkey: string(key)}
	if role == common.COMMITTEE {
		logger.Infof("Update peerID of MiningPubkey: %v %v", pid.String(), key)
		err := s.hc.chainData.UpdateCommittee(key, pid, byte(cID))
		if err != nil {
			return nil, err
		}

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
	logger.Infof("[blkbyhash] Receive GetBlockShardByHash request: %v %x", req.Shard, req.Hashes)
	defer s.reporter.watchRequestCounts("get_block_shard")

	// TODO(@0xbunyip): check if block in cache

	// Call node to get blocks
	// TODO(@0xbunyip): use fromPool
	data, err := s.hc.GetBlockShardByHash(
		req.Shard,
		req.Hashes,
	)
	if err != nil {
		logger.Infof("[blkbyhash] Receive GetBlockShardByHash response error: %v ", err)
		return nil, err
	}
	// TODO(@0xbunyip): cache blocks
	logger.Infof("[blkbyhash] Receive GetBlockShardByHash response data: %v ", data)
	return &proto.GetBlockShardByHashResponse{Data: data}, nil
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
	logger.Infof("Receive GetBlockBeaconByHash request: %x", req.Hashes)
	defer s.reporter.watchRequestCounts("get_block_beacon")

	// TODO(@0xbunyip): check if block in cache

	// Call node to get blocks
	// TODO(@0xbunyip): use fromPool
	data, err := s.hc.GetBlockBeaconByHash(
		req.Hashes,
	)
	if err != nil {
		return nil, err
	}
	// TODO(@0xbunyip): cache blocks
	return &proto.GetBlockBeaconByHashResponse{Data: data}, nil
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
