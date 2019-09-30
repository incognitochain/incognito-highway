package process

import (
	"context"
	fmt "fmt"
	logger "highway/customizelog"
	"highway/process/topic"

	"google.golang.org/grpc"
)

type GRPCService_Server struct {
}

func (self *GRPCService_Server) registerServices(grpsServer *grpc.Server) {
	RegisterProxyRegisterServiceServer(grpsServer, self)
}

func (self *GRPCService_Server) ProxyRegister(ctx context.Context, req *ProxyRegisterMsg) (*ProxyRegisterResponse, error) {
	logger.Infof("Receive new request from %v via gRPC\n", req.GetCommitteePublicKey())
	pairs := []*MessageTopicPair{}
	// topics := []string{}
	topicGenerator := new(topic.InsideTopic)
	for _, m := range req.WantedMessages {
		err := topicGenerator.FromMessageType(req.CommitteePublicKey, m)
		responseTopic := ""
		if err == nil {
			responseTopic = topicGenerator.ToString()
			logger.Infof("%v wanted message %v, response %v", req.GetCommitteePublicKey(), responseTopic)
			GlobalPubsub.NewMessage <- responseTopic
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
