package process

import (
	"context"
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"sync"

	"github.com/libp2p/go-libp2p-core/peer"
)

var CommitteePubkeyByPeerID map[peer.ID]string
var IsPublicIP map[peer.ID]bool
var updatePKbyPIDLocker sync.Mutex

func (s *HighwayServer) Register(
	ctx context.Context,
	req *RegisterRequest,
) (
	*RegisterResponse,
	error,
) {
	logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	pairs := []*MessageTopicPair{}
	var err error
	var pair *MessageTopicPair
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
		pairs = append(pairs, pair)
	}
	logger.Info(pairs)
	//	return &ProxyRegisterResponse{Pair: pairs}, nil

	// Notify HighwayClient of a new peer to request data later if possible
	pid, err := peer.IDB58Decode(req.PeerID)
	cid := common.GetCommitteeIDOfValidator(req.CommitteePublicKey)
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
	data, err := s.hc.GetBlockShardByHeight(
		req.Shard,
		req.FromHeight,
		req.ToHeight,
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
	return nil, nil
}

func (s *HighwayServer) GetBlockBeaconByHash(ctx context.Context, req *GetBlockBeaconByHashRequest) (*GetBlockBeaconByHashResponse, error) {
	logger.Println("Receive GetBlockBeaconByHash request")
	return nil, nil
}

func (s *HighwayServer) GetBlockCrossShardByHeight(ctx context.Context, req *GetBlockCrossShardByHeightRequest) (*GetBlockCrossShardByHeightResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHeight request")
	return nil, nil
}

func (s *HighwayServer) GetBlockCrossShardByHash(ctx context.Context, req *GetBlockCrossShardByHashRequest) (*GetBlockCrossShardByHashResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

type HighwayServer struct {
	hc *HighwayClient
}

func RunHighwayServer(h *p2p.Host, hc *HighwayClient) {
	s := &HighwayServer{
		hc: hc,
	}
	RegisterHighwayServiceServer(h.GRPC.GetGRPCServer(), s)
	go h.GRPC.Serve() // NOTE: must serve after registering all services
}
