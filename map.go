package main

import "github.com/libp2p/go-libp2p-core/peer"

type HighwayMap struct {
	peers   map[byte][]peer.AddrInfo // shard => peers
	support map[peer.ID][]byte       // peerID => shards supported

	connected []byte // keep track of connected shards
}

func NewHighwayMap(p peer.AddrInfo, supportShards []byte) *HighwayMap {
	cop := func(b []byte) []byte { // Create new slice and copy
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}

	m := &HighwayMap{
		peers:     map[byte][]peer.AddrInfo{},
		support:   map[peer.ID][]byte{p.ID: cop(supportShards)},
		connected: cop(supportShards),
	}

	for _, s := range supportShards {
		m.peers[s] = append(m.peers[s], p)
	}
	return m
}

func (h *HighwayMap) IsConnectedToShard(s byte) bool {
	for _, c := range h.connected {
		if c == s {
			return true
		}
	}
	return false
}

func (h *HighwayMap) ConnectToShard(s byte) {
	h.connected = append(h.connected, s)
}
