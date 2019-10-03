package main

import "github.com/libp2p/go-libp2p-core/peer"

type HighwayMap struct {
	Peers    map[byte][]peer.AddrInfo // shard => peers
	Supports map[peer.ID][]byte       // peerID => shards supported

	connected []byte // keep track of connected shards
}

func NewHighwayMap(p peer.AddrInfo, supportShards []byte) *HighwayMap {
	cop := func(b []byte) []byte { // Create new slice and copy
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}

	m := &HighwayMap{
		Peers:     map[byte][]peer.AddrInfo{},
		Supports:  map[peer.ID][]byte{p.ID: cop(supportShards)},
		connected: cop(supportShards),
	}

	for _, s := range supportShards {
		m.Peers[s] = append(m.Peers[s], p)
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

func (h *HighwayMap) ConnectToShardOfPeer(p peer.AddrInfo) {
	for _, s := range h.Supports[p.ID] {
		h.ConnectToShard(s)
	}
}
