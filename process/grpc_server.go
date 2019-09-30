package process

import (
	"context"
	"highway/process/topic"

	"google.golang.org/grpc"
)

var call = 0

type GRPCService_Server struct {
}

func (self *GRPCService_Server) registerServices(grpsServer *grpc.Server) {
	RegisterProxyRegisterServiceServer(grpsServer, self)
}

func (self *GRPCService_Server) ProxyRegister(ctx context.Context, req *ProxyRegisterMsg) (*ProxyRegisterResponse, error) {
	topics := []string{}
	topicGenerator := new(topic.InsideTopic)
	for _, m := range req.WantedMessages {
		err := topicGenerator.FromMessageType(req.CommitteePublicKey, m)
		if err != nil {
			topics = append(topics, "")
		} else {
			topics = append(topics, topicGenerator.ToString())
			GlobalPubsub.NewMessage <- topicGenerator.ToString()
		}
	}
	call += 1
	return &ProxyRegisterResponse{Topics: topics}, nil
}
