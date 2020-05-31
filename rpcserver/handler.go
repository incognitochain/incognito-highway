package rpcserver

import (
	"highway/common"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Handler struct {
	rpcServer *RpcServer
}

func (s *Handler) GetPeers(
	req Request,
	res *Response,
) (
	err error,
) {
	logger.Debugf("Received new GetPeers request: %+v", req)
	peers := s.rpcServer.pmap.CopyPeersMap()
	rpcs := s.rpcServer.pmap.CopyRPCUrls()
	status := s.rpcServer.pmap.CopyStatus()
	addrs := []HighwayAddr{}

	// NOTE: assume all highways support all shards => get at 0
	for _, p := range peers[0] {
		// Skip disconnected peers or just recently connected ones
		s, ok := status[p.ID]
		if !ok || !s.Connecting || time.Since(s.Start) < common.MinStableDuration {
			continue
		}

		ma, err := peer.AddrInfoToP2pAddrs(&p)
		if err != nil {
			logger.Warnf("Invalid addr info: %+v", p)
			continue
		}

		nonLocal := common.FilterLocalAddrs(ma)
		addr := ma[0].String()
		if len(nonLocal) > 0 {
			addr = nonLocal[0].String()
		}
		addrs = append(addrs, HighwayAddr{
			Libp2pAddr: addr,
			RPCUrl:     rpcs[p.ID],
		})
	}

	res.PeerPerShard = map[string][]HighwayAddr{
		"all": addrs,
	}
	logger.Debugf("GetPeers return: %+v", res.PeerPerShard)
	return
}
