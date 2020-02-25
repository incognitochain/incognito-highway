package chain

import (
	"context"
	"highway/chaindata"
	"highway/common"
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
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
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
	r := chaindata.GetUserRole(reqRole, cID)
	pid, err := peer.IDB58Decode(req.PeerID)
	if err != nil {
		logger.Warnf("Invalid peerID: %v", req.PeerID)
		return nil, errors.WithStack(err)
	}

	key, err := common.PreprocessKey(req.GetCommitteePublicKey())
	if err != nil {
		return nil, err
	}

	pinfo := PeerInfo{ID: pid, Pubkey: string(key)}
	if role == common.COMMITTEE {
		logger.Infof("Update peerID of MiningPubkey: %v %v", pid.String(), key)
		s.chainData.UpdateCommittee(key, pid, byte(cID))
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

// req requestByHeight, heights []uint64
func (s *Server) GetBlockByHeight(ctx context.Context, req getBlockByHeightRequest) ([][]byte, error) {
	if req.GetCallDepth() > common.MaxCallDepth {
		err := errors.Errorf("reached max call depth: %+v", req)
		return nil, err
	}
	heights := convertToSpecificHeights(req.GetSpecific(), req.GetFromHeight(), req.GetToHeight(), req.GetHeights())
	idxs := make([]int, len(heights))
	for i := 0; i < len(idxs); i++ {
		idxs[i] = i
	}

	blocks := make([][]byte, len(heights))
	for _, p := range s.providers {
		if len(heights) == 0 {
			break
		}

		data, err := p.GetBlockByHeight(ctx, req, heights)
		if err != nil {
			logger.Warnf("Failed GetBlockByHeight: %+v", err)
			continue
		}

		newHeights := []uint64{}
		newIdxs := []int{}
		for i, d := range data {
			if d == nil {
				// Nil result, must ask next provider
				newHeights = append(newHeights, heights[i])
				newIdxs = append(newIdxs, idxs[i])
				continue
			}

			blocks[idxs[i]] = d
		}
		heights = newHeights
		idxs = newIdxs
	}
	return blocks, nil
}

func (s *Server) GetBlockShardByHeight(ctx context.Context, req *proto.GetBlockShardByHeightRequest) (*proto.GetBlockShardByHeightResponse, error) {
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)

	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_shard")

	// Call node to get blocks
	data, err := s.GetBlockByHeight(ctx, req)
	if err != nil {
		logger.Warnf("GetBlockShardByHeight return error: %+v", err)
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &proto.GetBlockShardByHeightResponse{Data: data}, nil
}

func (s *Server) GetBlockBeaconByHeight(ctx context.Context, req *proto.GetBlockBeaconByHeightRequest) (*proto.GetBlockBeaconByHeightResponse, error) {
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)

	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_beacon")

	// Call node to get blocks
	data, err := s.GetBlockByHeight(ctx, req)
	if err != nil {
		logger.Warnf("GetBlockBeaconByHeight return error: %+v", err)
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
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)

	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_shard_to_beacon")

	data, err := s.GetBlockByHeight(ctx, req)
	if err != nil {
		logger.Warnf("GetBlockShardToBeaconByHeight return error: %+v", err)
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &proto.GetBlockShardToBeaconByHeightResponse{Data: data}, nil
}

func (s *Server) GetBlockCrossShardByHeight(ctx context.Context, req *proto.GetBlockCrossShardByHeightRequest) (*proto.GetBlockCrossShardByHeightResponse, error) {
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)

	// Monitor status
	defer s.reporter.watchRequestCounts("get_block_cross_shard")

	data, err := s.GetBlockByHeight(ctx, req)
	if err != nil {
		logger.Warnf("GetBlockCrossShardByHeight return error: %+v", err)
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &proto.GetBlockCrossShardByHeightResponse{Data: data}, nil
}

func convertToSpecificHeights(
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) []uint64 {
	to, heights = capBlocksPerRequest(specific, from, to, heights)
	if !specific {
		heights = make([]uint64, to-from+1)
		for i := from; i <= to; i++ {
			heights[i-from] = i
		}
	}
	return heights
}

func (s *Server) GetBlockByHash(ctx context.Context, req getBlockByHashRequest) ([][]byte, error) {
	if req.GetCallDepth() > common.MaxCallDepth {
		err := errors.Errorf("reached max call depth: %+v", req)
		return nil, err
	}
	hashes := req.GetHashes()
	idxs := make([]int, len(hashes))
	for i := 0; i < len(idxs); i++ {
		idxs[i] = i
	}

	blocks := make([][]byte, len(hashes))
	for _, p := range s.providers {
		if len(hashes) == 0 {
			break
		}

		data, err := p.GetBlockByHash(ctx, req, hashes)
		if err != nil {
			logger.Warnf("Failed GetBlockByHash: %+v", err)
			continue
		}

		newHashes := [][]byte{}
		newIdxs := []int{}
		for i, d := range data {
			if d == nil {
				// Nil result, must ask next provider
				newHashes = append(newHashes, hashes[i])
				newIdxs = append(newIdxs, idxs[i])
				continue
			}

			blocks[idxs[i]] = d
		}
		hashes = newHashes
		idxs = newIdxs
	}
	return blocks, nil
}

func (s *Server) GetBlockShardByHash(ctx context.Context, req *proto.GetBlockShardByHashRequest) (*proto.GetBlockShardByHashResponse, error) {
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)

	logger.Infof("[blkbyhash] Receive GetBlockShardByHash request: %v %x", req.Shard, req.Hashes)
	defer s.reporter.watchRequestCounts("get_block_shard")

	data, err := s.GetBlockByHash(ctx, req)
	if err != nil {
		logger.Warnf("GetBlockShardByHash return error: %+v", err)
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	logger.Infof("[blkbyhash] Receive GetBlockShardByHash response data: %v ", data)
	return &proto.GetBlockShardByHashResponse{Data: data}, nil
}

func (s *Server) GetBlockBeaconByHash(ctx context.Context, req *proto.GetBlockBeaconByHashRequest) (*proto.GetBlockBeaconByHashResponse, error) {
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Infof("Receive GetBlockBeaconByHash request: %x", req.Hashes)
	defer s.reporter.watchRequestCounts("get_block_beacon")

	data, err := s.GetBlockByHash(ctx, req)
	if err != nil {
		logger.Warnf("GetBlockBeaconByHash return error: %+v", err)
		return nil, err
	}

	// TODO(@0xbunyip): cache blocks
	return &proto.GetBlockBeaconByHashResponse{Data: data}, nil
}

func (s *Server) GetBlockCrossShardByHash(ctx context.Context, req *proto.GetBlockCrossShardByHashRequest) (*proto.GetBlockCrossShardByHashResponse, error) {
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Errorf("Receive GetBlockCrossShardByHash request: %d %d %x", req.FromShard, req.ToShard, req.Hashes)
	return nil, errors.New("not supported")
}

type Server struct {
	proto.UnimplementedHighwayServiceServer
	m         *Manager
	providers []Provider
	chainData *chaindata.ChainData

	reporter *Reporter
	// blkgetter BlockGetter
}

type Provider interface {
	GetBlockByHeight(ctx context.Context, req getBlockByHeightRequest, heights []uint64) ([][]byte, error)
	GetBlockByHash(ctx context.Context, req getBlockByHashRequest, hashes [][]byte) ([][]byte, error)
}

func RegisterServer(
	m *Manager,
	gs *grpc.Server,
	hc *Client,
	chainData *chaindata.ChainData,
	reporter *Reporter,
) {
	s := &Server{
		providers: []Provider{hc},
		m:         m,
		reporter:  reporter,
		chainData: chainData,
	}
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

// capBlocksPerRequest returns the maximum height allowed for a single request
// If the request is for a range, this function returns the maximum block height allowed
// If the request is for some blocks, this caps the number blocks requested
func capBlocksPerRequest(specific bool, from, to uint64, heights []uint64) (uint64, []uint64) {
	if specific {
		if len(heights) > common.MaxBlocksPerRequest {
			heights = heights[:common.MaxBlocksPerRequest]
		}
		return heights[len(heights)-1], heights
	}

	maxHeight := from + common.MaxBlocksPerRequest
	if to > maxHeight {
		return maxHeight, heights
	}
	return to, heights
}
