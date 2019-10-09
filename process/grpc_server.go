package process

import (
	"context"
	fmt "fmt"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process/topic"

	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

func ProcessConnection(h *p2p.Host) {
	s := &HighwayServer{}
	RegisterHighwayServiceServer(h.GRPC.GetGRPCServer(), s)
}

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

func (s *HighwayServer) GetBlockShardByHeight(context.Context, *GetBlockShardByHeightRequest) (*GetBlockShardByHeightResponse, error) {
	logger.Println("Receive GetBlockShardByHeight request")
	return nil, nil
}

func (s *HighwayServer) GetBlockShardByHash(context.Context, *GetBlockShardByHashRequest) (*GetBlockShardByHashResponse, error) {
	logger.Println("Receive GetBlockShardByHash request")
	return nil, nil
}

func (s *HighwayServer) GetBlockBeaconByHeight(context.Context, *GetBlockBeaconByHeightRequest) (*GetBlockBeaconByHeightResponse, error) {
	logger.Println("Receive GetBlockBeaconByHeight request")
	return nil, nil
}

func (s *HighwayServer) GetBlockBeaconByHash(context.Context, *GetBlockBeaconByHashRequest) (*GetBlockBeaconByHashResponse, error) {
	logger.Println("Receive GetBlockBeaconByHash request")
	return nil, nil
}

func (s *HighwayServer) GetBlockCrossShardByHeight(context.Context, *GetBlockCrossShardByHeightRequest) (*GetBlockCrossShardByHeightResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHeight request")
	return nil, nil
}

func (s *HighwayServer) GetBlockCrossShardByHash(context.Context, *GetBlockCrossShardByHashRequest) (*GetBlockCrossShardByHashResponse, error) {
	logger.Println("Receive GetBlockCrossShardByHash request")
	return nil, nil
}

type HighwayServer struct{}
