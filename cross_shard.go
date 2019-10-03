package main

import (
	"highway/common"
	logger "highway/customizelog"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/stathat/consistent"
)

type Highway struct {
	SupportShards []byte
	ID            peer.ID

	hmap *HighwayMap
	hc   *HighwayConnector
}

func NewHighway(
	supportShards []byte,
	bootstrap []string,
	h host.Host,
) *Highway {
	// TODO(@0xbunyip): use bootstrap to get initial highways
	p := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	hmap := NewHighwayMap(p, supportShards)

	if len(bootstrap) > 0 && len(bootstrap[0]) > 0 {
		ss := []byte{0}
		id, _ := peer.IDB58Decode("12D3KooW9sbQK4J64Qat5D9vQEhBCnDYD3WPqWmgUZD4M7CJ2rXS")
		addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9330")
		hmap.AddPeer(peer.AddrInfo{
			ID:    id,
			Addrs: []multiaddr.Multiaddr{addr},
		}, ss)
	}

	hw := &Highway{
		SupportShards: supportShards,
		ID:            h.ID(),
		hmap:          hmap,
		hc:            NewHighwayConnector(h, hmap),
	}
	go hw.hc.Start()
	return hw
}

func (h *Highway) Start() {
	for range time.Tick(5 * time.Second) {
		// TODO(@0xbunyip) Check for liveness of connected highways

		// New highways online: update map and reconnect to load-balance
		newHighway := true
		_ = newHighway

		// Connect to other highways if needed
		h.UpdateConnection()
	}
}

func (h *Highway) UpdateConnection() {
	for i := byte(0); i < common.NumberOfShard; i++ {
		if err := h.connectChain(i); err != nil {
			logger.Error(err)
		}
	}

	if err := h.connectChain(common.BEACONID); err != nil {
		logger.Error(err)
	}
}

// connectChain connects this highway to a peer in a chain (shard or beacon) if it hasn't connected to one yet
func (h *Highway) connectChain(sid byte) error {
	if h.hmap.IsConnectedToShard(sid) {
		return nil
	}

	logger.Info("connect chain", sid)
	highways := h.hmap.Peers[sid]
	if len(highways) == 0 {
		return errors.Errorf("found no highway supporting shard %d", sid)
	}

	// TODO(@0xbunyip): repick if fail to connect
	p, err := choosePeer(highways, h.ID)
	if err != nil {
		return errors.WithMessagef(err, "shardID: %v", sid)
	}
	if err := h.connectTo(p); err != nil {
		return err
	}

	// Update list of connected shards
	h.hmap.ConnectToShardOfPeer(p)
	return nil
}

func (h *Highway) connectTo(p peer.AddrInfo) error {
	return h.hc.ConnectTo(p)
}

// choosePeer picks a peer from a list using consistent hashing
func choosePeer(peers []peer.AddrInfo, id peer.ID) (peer.AddrInfo, error) {
	cst := consistent.New()
	for _, p := range peers {
		cst.Add(string(p.ID))
	}

	closest, err := cst.Get(string(id))
	if err != nil {
		return peer.AddrInfo{}, errors.New("could not get consistent-hashing peer")
	}

	for _, p := range peers {
		if string(p.ID) == closest {
			return p, nil
		}
	}
	return peer.AddrInfo{}, errors.New("failed choosing peer to connect")
}
