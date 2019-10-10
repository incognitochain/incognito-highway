package process

import (
	"context"
	fmt "fmt"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process/topic"

	"github.com/libp2p/go-libp2p-core/peer"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

func (s *HighwayServer) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	logger.Infof("Receive new request from %v via gRPC", req.GetCommitteePublicKey())
	pairs := []*MessageTopicPair{}
	topicGenerator := new(topic.InsideTopic)
	for _, m := range req.WantedMessages {
		err := topicGenerator.FromMessageType(req.CommitteePublicKey, m)
		responseTopic := ""
		if err == nil {
			responseTopic = topicGenerator.ToString()
			logger.Infof("Someone wanted message %v, response %v", m, responseTopic)
			GlobalPubsub.NewMessage <- SubHandler{
				Topic:   responseTopic,
				Handler: func(*p2pPubSub.Message) {},
			}
		}
		pair := &MessageTopicPair{
			Message: m,
			Topic:   responseTopic,
		}
		pairs = append(pairs, pair)
	}
	fmt.Println(pairs)
	return &RegisterResponse{Pair: pairs}, nil
}

func (s *HighwayServer) GetBlockShardByHeight(ctx context.Context, req *GetBlockShardByHeightRequest) (*GetBlockShardByHeightResponse, error) {
	logger.Println("Receive GetBlockShardByHeight request")
	// TODO(@0xbunyip): check if block in cache
	// TODO(0xakk0r0kamui): choose client from peer state
	var peerID peer.ID

	// Call node to get blocks
	data, err := s.hc.GetBlockShardByHeight(
		peerID,
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
}
