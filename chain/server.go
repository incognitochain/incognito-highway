package chain

import (
	"context"
	"highway/chaindata"
	"highway/common"
	"highway/process"
	"highway/process/datahandler"
	"highway/process/topic"
	"highway/proto"

	"github.com/patrickmn/go-cache"
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
	pairs, err := s.processListWantedMessageOfPeer(req.GetWantedMessages(), role, cIDs, req.CommitteePublicKey)
	if err != nil {
		logger.Warnf("Couldn't process wantedMsgs: %+v %+v %+v", req.GetWantedMessages(), role, cIDs)
		return nil, err
	}

	for _, p := range pairs {
		for _, t := range p.Topic {
			msgType := topic.GetMsgTypeOfTopic(t)
			if _, ok := common.TopicPrivate[msgType]; ok {
				if common.HasStringAt(s.psManager.FollowedTopic, t) == 0 {
					continue
				}
				handler := datahandler.SubsHandler{
					PubSub:         s.psManager.FloodMachine,
					FromInside:     true,
					BlockchainData: s.psManager.BlockChainData,
					Cacher:         cache.New(common.MaxTimeKeepPubSubData, common.MaxTimeKeepPubSubData),
					CommitteeInfo:  s.psManager.CommitteeInfo,
					Scenario:       s.psManager.Scenario,
				}
				go func() {
					subs, err := s.psManager.FloodMachine.Subscribe(t)
					if err != nil {
						logger.Info(err)
						return
					}
					logger.Infof("Success subscribe topic %v", t)
					logger.Infof("aaaaaaaaaaaaaaa %v", handler.Scenario)
					logger.Infof("aaaaaaaaaaaaaaa %v", handler.Scenario.Lock)
					err = handler.HandlerNewSubs(subs)
					if err != nil {
						logger.Errorf("Handle Subsciption topic %v return error %v", subs.Topic(), err)
					}
				}()
				s.psManager.FollowedTopic = append(s.psManager.FollowedTopic, t)
			}
		}
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
		s.psManager.CommitteeInfo.AddPubKey(string(key), byte(cID))
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

func (s *Server) GetBlockByHash(ctx context.Context, req GetBlockByHashRequest) ([][]byte, error) {
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
	for _, p := range s.Providers {
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
	Providers []Provider
	chainData *chaindata.ChainData

	psManager *process.PubSubManager

	reporter *Reporter
	// blkgetter BlockGetter
}

type Provider interface {
	SetBlockByHeight(ctx context.Context, req GetBlockByHeightRequest, heights []uint64, blocks [][]byte) error
	// GetBlockByHeight(ctx context.Context, req GetBlockByHeightRequest, heights []uint64) ([][]byte, error)
	GetBlockByHash(ctx context.Context, req GetBlockByHashRequest, hashes [][]byte) ([][]byte, error)
	StreamBlkByHeight(ctx context.Context, req RequestBlockByHeight, blkChan chan common.ExpectedBlkByHeight) error
	StreamBlkByHeightv2(ctx context.Context, req RequestBlockByHeight, blkChan chan common.ExpectedBlk) error
	StreamBlkByHash(ctx context.Context, req RequestBlockByHash, blkChan chan common.ExpectedBlk) error
	SetSingleBlockByHeight(ctx context.Context, req RequestBlockByHeight, data common.ExpectedBlkByHeight) error
	SetSingleBlockByHash(ctx context.Context, req RequestBlockByHash, data common.ExpectedBlk) error
	SetSingleBlockByHeightv2(ctx context.Context, req RequestBlockByHeight, data common.ExpectedBlk) error
}

func RegisterServer(
	m *Manager,
	gs *grpc.Server,
	hc *Client,
	chainData *chaindata.ChainData,
	reporter *Reporter,
	psManager *process.PubSubManager,
) (*Server, error) {
	memcache, err := NewRistrettoMemCache()
	if err != nil {
		return nil, err
	}

	s := &Server{
		Providers: []Provider{memcache, hc}, // NOTE: memcache must go before client
		m:         m,
		reporter:  reporter,
		chainData: chainData,
		psManager: psManager,
	}
	proto.RegisterHighwayServiceServer(gs, s)
	return s, nil
}

func (s *Server) processListWantedMessageOfPeer(
	msgs []string,
	role byte,
	committeeIDs []int,
	peerPubKey string,
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
	pairs = topic.Handler.GetListTopicPairForNode(role, msgAndCID, peerPubKey)
	return pairs, nil
}

// capBlocksPerRequest returns the maximum height allowed for a single request
// If the request is for a range, this function returns the maximum block height allowed
// If the request is for some blocks, this caps the number blocks requested
func capBlocksPerRequest(specific bool, from, to uint64, heights []uint64) (uint64, []uint64) {
	if specific {
		if uint64(len(heights)) > common.MaxBlocksPerRequest {
			heights = heights[:common.MaxBlocksPerRequest]
		}
		return heights[len(heights)-1], heights
	}

	maxHeight := from + common.MaxBlocksPerRequest - 1
	if to > maxHeight {
		return maxHeight, heights
	}
	return to, heights
}
