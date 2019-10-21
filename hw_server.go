package main

import (
	"context"
	"highway/process"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
)

func (s *HWServer) GetChainCommittee(ctx context.Context, req *process.GetChainCommitteeRequest) (*process.GetChainCommitteeResponse, error) {
	// TODO(@0xbunyip): get ChainCommittee from ChainData and return here
	return &process.GetChainCommitteeResponse{Data: make([]byte, 3)}, nil
}

func NewHWServer(prtc *p2pgrpc.GRPCProtocol) *HWServer {
	s := &HWServer{}
	process.RegisterHighwayConnectorServiceServer(prtc.GetGRPCServer(), s)
	return s
}

type HWServer struct{}
