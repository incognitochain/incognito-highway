package rpcserver

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type Handler struct {
	rpcServer *RpcServer
}

type PeerMap interface {
	CopyPeersMap() map[byte][]peer.AddrInfo
}

func (s *Handler) GetPeers(
	req Request,
	res *Response,
) (
	err error,
) {
	peers := s.rpcServer.pmap.CopyPeersMap()
	addrs := []string{}

	// NOTE: assume all highways support all shards => get at 0
	for _, p := range peers[0] {
		// TODO(@0xbunyip): get p[idx] with public ip
		ma, err := peer.AddrInfoToP2pAddrs(&p)
		if err != nil {
			logger.Warnf("Invalid addr info: %+v", p)
			continue
		}

		nonLocal := filterLocalAddrs(ma)
		addr := ma[0].String()
		if len(nonLocal) > 0 {
			addr = nonLocal[0].String()
		}
		addrs = append(addrs, addr)
	}

	res.PeerPerShard = map[string][]string{
		"all": addrs,
	}
	return
}

func filterLocalAddrs(mas []ma.Multiaddr) []ma.Multiaddr {
	localAddrs := []string{
		"127.0.0.1",
		"0.0.0.0",
	}
	nonLocal := []ma.Multiaddr{}
	for _, ma := range mas {
		local := false
		for _, s := range localAddrs {
			if strings.Contains(ma.String(), s) {
				local = true
				break
			}
		}
		if !local {
			nonLocal = append(nonLocal, ma)
		}
	}
	fmt.Println(mas, nonLocal)
	return nonLocal
}
