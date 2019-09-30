package process

import (
	"highway/p2p"
)

func ProcessConnection(s *p2p.Host) {
	gRPCService := GRPCService_Server{}
	gRPCService.registerServices(s.GRPC.GetGRPCServer())
}
