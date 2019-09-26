package main

import (
	"context"

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
	for i, m := range req.WantedMessages {
		topic := "PROXY" + m
		// if call%10 == i {
		// 	topic = topic[1:]
		// }
		topics = append(topics, topic)
	}
	call += 1
	return &ProxyRegisterResponse{Topics: topics}, nil
}
