package chain

import (
	"context"
	"highway/common"
	logger "highway/customizelog"
	"highway/process"
	"highway/proto"

	"github.com/libp2p/go-libp2p-core/peer"
	grpc "google.golang.org/grpc"
)

func (s *Server) Register(
	ctx context.Context,
	req *proto.RegisterRequest,
) (
	*proto.RegisterResponse,
	error,
) {
	logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	pairs := []*proto.MessageTopicPair{}
	var err error
	var pair *proto.MessageTopicPair
	for _, m := range req.WantedMessages {
		pair, err = process.GenerateResponseTopic(&process.GlobalPubsub, req.CommitteePublicKey, m)
		if err != nil {
			logger.Infof("GenerateResponseTopic failed, error: %v", err.Error())
			pair = &proto.MessageTopicPair{
				Message: m,
				Topic:   []string{},
				Act:     []proto.MessageTopicPair_Action{},
			}
		}
		peerid, _ := peer.IDB58Decode(req.GetPeerID())
		process.UpdatePeerIDOfCommitteePubkey(req.GetCommitteePublicKey(), &peerid)
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
	logger.Println("Receive GetBlockShardByHash request")
	return nil, nil
}

func (s *Server) GetBlockBeaconByHeight(ctx context.Context, req *proto.GetBlockBeaconByHeightRequest) (*proto.GetBlockBeaconByHeightResponse, error) {
	// logger.Println("Receive GetBlockBeaconByHeight request")
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

func (s *Server) GetBlockBeaconByHash(ctx context.Context, req *proto.GetBlockBeaconByHashRequest) (*proto.GetBlockBeaconByHashResponse, error) {
	logger.Println("Receive GetBlockBeaconByHash request")
	return nil, nil
}

func (s *Server) GetBlockCrossShardByHeight(ctx context.Context, req *proto.GetBlockCrossShardByHeightRequest) (*proto.GetBlockCrossShardByHeightResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHeight request")
	return nil, nil
}

func (s *Server) GetBlockCrossShardByHash(ctx context.Context, req *proto.GetBlockCrossShardByHashRequest) (*proto.GetBlockCrossShardByHashResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

type Server struct {
	hc *Client
}

func RegisterServer(gs *grpc.Server, hc *Client) {
	s := &Server{hc: hc}
	proto.RegisterHighwayServiceServer(gs, s)
}
