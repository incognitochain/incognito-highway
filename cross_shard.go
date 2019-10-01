package main

import (
	"highway/common"
	logger "highway/customizelog"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/stathat/consistent"
)

type HighwayManager struct {
	SupportShards []byte
	ID            peer.ID

	shardsConnected []byte
}

func NewHighwayManager(supportShards []byte, bootstrap []string, id peer.ID) *HighwayManager {
	// TODO(@0xbunyip): use bootstrap to get initial highways
	sc := make([]byte, len(supportShards))
	copy(sc, supportShards)
	hm := &HighwayManager{
		SupportShards:   supportShards,
		ID:              id,
		shardsConnected: sc,
	}
	return hm
}

func (hm *HighwayManager) Start() {
	for range time.Tick(60 * time.Second) {
		// TODO(@0xbunyip) Check for liveness of connected highways

		// New highways online: update map and reconnect to load-balance
		newHighway := true
		_ = newHighway

		// Connect to other highways if needed
		highwayMap := map[byte][]peer.AddrInfo{}
		hm.UpdateConnection(highwayMap)
	}
}

func (hm *HighwayManager) UpdateConnection(highways map[byte][]peer.AddrInfo) {
	for i := byte(0); i < common.NumberOfShard; i++ {
		if err := hm.connectChain(highways, i); err != nil {
			logger.Error(err)
		}
	}

	if err := hm.connectChain(highways, common.BEACONID); err != nil {
		logger.Error(err)
	}
}

// connectChain connects this highway to a chain (shard or beacon) if it hasn't connected to one yet
func (hm *HighwayManager) connectChain(highways map[byte][]peer.AddrInfo, sid byte) error {
	if contain(sid, hm.shardsConnected) { // Connected
		return nil
	}

	if len(highways[sid]) == 0 {
		return errors.Errorf("Found no highway supporting shard %d", sid)
	}

	p, err := choosePeer(highways[sid], hm.ID)
	if err != nil {
		return errors.WithMessagef(err, "shardID: %v", sid)
	}
	if err := hm.connectTo(p); err != nil {
		return err
	}

	// Update list of connected shards
	// TODO(@0xbunyip): update more than one sid when highway supports many
	hm.shardsConnected = append(hm.shardsConnected, sid)
	return nil
}

func (hm *HighwayManager) connectTo(p peer.AddrInfo) error {
	return nil
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
