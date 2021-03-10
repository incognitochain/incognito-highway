package route

import (
	"context"
	"highway/proto"
	hmap "highway/route/hmap"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func (s *Server) GetChainCommittee(ctx context.Context, req *proto.GetChainCommitteeRequest) (*proto.GetChainCommitteeResponse, error) {
	// TODO(@0xbunyip): get ChainCommittee from ChainData and return here
	return &proto.GetChainCommitteeResponse{Data: make([]byte, 3)}, nil
}

func (s *Server) GetHighwayInfos(ctx context.Context, req *proto.GetHighwayInfosRequest) (*proto.GetHighwayInfosResponse, error) {
	peers := s.hmap.CopyPeersMap()
	resp := &proto.GetHighwayInfosResponse{Highways: []*proto.HighwayInfo{}}
	for sid, addrs := range peers {
		for _, addr := range addrs {
			ma, err := peer.AddrInfoToP2pAddrs(&addr)
			if err != nil {
				continue
			}

			found := false
			pinfo := ma[0].String()
			for _, hi := range resp.Highways {
				if hi.PeerInfo == pinfo {
					hi.SupportShards = append(hi.SupportShards, int32(sid))
					found = true
					break
				}
			}
			if found {
				continue
			}

			hi := &proto.HighwayInfo{PeerInfo: pinfo, SupportShards: []int32{int32(sid)}}
			resp.Highways = append(resp.Highways, hi)
		}
	}
	return resp, nil
}

func NewServer(prtc *p2pgrpc.GRPCProtocol, hmap *hmap.Map) *Server {
	s := &Server{hmap: hmap}
	sHealth := health.NewServer()
	proto.RegisterHighwayConnectorServiceServer(prtc.GetGRPCServer(), s)
	healthpb.RegisterHealthServer(prtc.GetGRPCServer(), sHealth)
	return s
}

type Server struct {
	hmap *hmap.Map
}
