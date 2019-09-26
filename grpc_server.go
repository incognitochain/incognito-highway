package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

type GRPCService_Server struct {
}

func (self *GRPCService_Server) registerServices(grpsServer *grpc.Server) {
	RegisterProxyRegisterServiceServer(grpsServer, self)
}

func (self *GRPCService_Server) ProxyRegister(ctx context.Context, req *ProxyRegisterMsg) (*ProxyRegisterResponse, error) {
	fmt.Println(req)
	return &ProxyRegisterResponse{Topics: []string{"123", "456", "789"}}, nil
}
