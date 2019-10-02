package main

import (
	"highway/common"
	logger "highway/customizelog"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	host "github.com/libp2p/go-libp2p-host"
	"github.com/pkg/errors"
	"github.com/stathat/consistent"
)

type Highway struct {
	SupportShards []byte
	ID            peer.ID

	shardsConnected []byte
	hc              *HighwayConnector
}

func NewHighway(
	supportShards []byte,
	bootstrap []string,
	h host.Host,
) *Highway {
	// TODO(@0xbunyip): use bootstrap to get initial highways
	sc := make([]byte, len(supportShards))
	copy(sc, supportShards)
	hw := &Highway{
		SupportShards:   supportShards,
		ID:              h.ID(),
		shardsConnected: sc,
		hc:              NewHighwayConnector(h),
	}
	return hw
}

func (h *Highway) Start() {
	for range time.Tick(60 * time.Second) {
		// TODO(@0xbunyip) Check for liveness of connected highways

		// New highways online: update map and reconnect to load-balance
		newHighway := true
		_ = newHighway

		// Connect to other highways if needed
		highwayMap := map[byte][]peer.AddrInfo{}
		h.UpdateConnection(highwayMap)
	}
}

func (h *Highway) UpdateConnection(highways map[byte][]peer.AddrInfo) {
	for i := byte(0); i < common.NumberOfShard; i++ {
		if err := h.connectChain(highways, i); err != nil {
			logger.Error(err)
		}
	}

	if err := h.connectChain(highways, common.BEACONID); err != nil {
		logger.Error(err)
	}
}

// connectChain connects this highway to a chain (shard or beacon) if it hasn't connected to one yet
func (h *Highway) connectChain(highways map[byte][]peer.AddrInfo, sid byte) error {
	if contain(sid, h.shardsConnected) { // Connected
		return nil
	}

	if len(highways[sid]) == 0 {
		return errors.Errorf("Found no highway supporting shard %d", sid)
	}

	p, err := choosePeer(highways[sid], h.ID)
	if err != nil {
		return errors.WithMessagef(err, "shardID: %v", sid)
	}
	if err := h.connectTo(p); err != nil {
		return err
	}

	// Update list of connected shards
	// TODO(@0xbunyip): update more than one sid when highway supports many
	h.shardsConnected = append(h.shardsConnected, sid)
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

func contain(b byte, l []byte) bool {
	for _, m := range l {
		if b == m {
			return true
		}
	}
	return false
}
