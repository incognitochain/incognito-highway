package route

import (
	"context"
	"highway/proto"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
)

func (s *Server) GetChainCommittee(ctx context.Context, req *proto.GetChainCommitteeRequest) (*proto.GetChainCommitteeResponse, error) {
	// TODO(@0xbunyip): get ChainCommittee from ChainData and return here
	return &proto.GetChainCommitteeResponse{Data: make([]byte, 3)}, nil
}

func NewServer(prtc *p2pgrpc.GRPCProtocol) *Server {
	s := &Server{}
	proto.RegisterHighwayConnectorServiceServer(prtc.GetGRPCServer(), s)
	return s
}

type Server struct{}
