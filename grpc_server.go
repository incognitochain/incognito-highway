package main

import (
	"context"
	"google.golang.org/grpc"
)

type GRPCService_Server struct {
}

func (self *GRPCService_Server) registerServices(grpsServer *grpc.Server) {
	RegisterProxyRegisterServiceServer(grpsServer, self)
}

func (self *GRPCService_Server) ProxyRegister(ctx context.Context, req *ProxyRegisterMsg) (*ProxyRegisterResponse, error) {
	return &ProxyRegisterResponse{Result: "ok nha"}, nil
}
