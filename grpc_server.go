package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

var call = 0

type GRPCService_Server struct {
}

func (self *GRPCService_Server) registerServices(grpsServer *grpc.Server) {
	RegisterProxyRegisterServiceServer(grpsServer, self)
}

func (self *GRPCService_Server) ProxyRegister(ctx context.Context, req *ProxyRegisterMsg) (*ProxyRegisterResponse, error) {
	pairs := []*MessageTopicPair{}
	for _, m := range req.WantedMessages {
		pair := &MessageTopicPair{
			Message: m,
			Topic:   "PROXY" + m,
		}
		// if call%10 == i {
		// 	pair.Topic = pair.Topic[1:]
		// }
		pairs = append(pairs, pair)
	}
	call += 1
	fmt.Println(pairs)
	return &ProxyRegisterResponse{Pair: pairs}, nil
}
