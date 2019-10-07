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
	s := &GRPCService_Server{}
	RegisterHighwayServiceServer(h.GRPC.GetGRPCServer(), s)
}

func (s *GRPCService_Server) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
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
	return &ProxyRegisterResponse{Pair: pairs}, nil
}

type GRPCService_Server struct{}
