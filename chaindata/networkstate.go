package chaindata

import (
	"encoding/json"
	"fmt"
	"highway/common"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type ChainState struct {
	Timestamp     int64
	Height        uint64
	BlockHash     [32]byte
	BestStateHash [32]byte
}

func (m ChainState) MarshalJSON() ([]byte, error) {
	pretty := map[string]interface{}{
		"Timestamp":     m.Timestamp,
		"Height":        m.Height,
		"BlockHash":     fmt.Sprintf("%x", m.BlockHash[:]),
		"BestStateHash": fmt.Sprintf("%x", m.BestStateHash[:]),
	}
	return json.Marshal(pretty)
}

type CommitteeState map[string][]byte

// NetworkState contains all of chainstate of node, which in committee and connected to proxy
type NetworkState struct {
	BeaconState          map[string]ChainState // map[<Committee Public Key>]Chainstate
	beaconLocker         *sync.RWMutex
	ShardState           map[byte]map[string]ChainState // map[<ShardID>]map[<Committee Public Key>]Chainstate
	shardLocker          *sync.RWMutex
	highwayIDOfPublicKey *cache.Cache
}

func (nwState *NetworkState) Init(numberOfShard int) {
	nwState.beaconLocker = new(sync.RWMutex)
	nwState.BeaconState = map[string]ChainState{}
	nwState.shardLocker = new(sync.RWMutex)
	nwState.ShardState = map[byte]map[string]ChainState{}
	for i := byte(0); i < byte(numberOfShard); i++ {
		nwState.ShardState[i] = map[string]ChainState{}
	}
	nwState.highwayIDOfPublicKey = cache.New(common.MaxTimeKeepPeerState, common.MaxTimeKeepPeerState)
	nwState.highwayIDOfPublicKey.OnEvicted(nwState.DeletePeerInfo)
}

func (nwState *NetworkState) GetHWIDOfPubKey(
	pubKey string,
) (
	peer.ID,
	error,
) {
	hwIDFromCache, ok := nwState.highwayIDOfPublicKey.Get(pubKey)
	if (!ok) || (hwIDFromCache == nil) {
		return "", errors.Errorf("Can not found highway ID for pubkey %v", pubKey)
	}
	return hwIDFromCache.(peer.ID), nil
}

func (nwState *NetworkState) GetAllHWIDInfo() map[string]string {
	hwInfo := map[string]string{}
	hwInfoCached := nwState.highwayIDOfPublicKey.Items()
	for pk, peerCached := range hwInfoCached {
		if peerCached.Object == nil {
			continue
		}
		pID := peerCached.Object.(peer.ID)
		hwInfo[pk] = pID.String()
	}
	return hwInfo
}

func (nwState *NetworkState) SetHWIDOfPubKey(
	hwID peer.ID,
	pubKey string,
) error {
	nwState.highwayIDOfPublicKey.Set(pubKey, hwID, common.MaxTimeKeepPeerState)
	return nil
}

func (nwState *NetworkState) DeletePeerInfo(peerPK string, highwayID interface{}) {
	logger.Infof("[delpeerstate] key %v", peerPK)
	nwState.beaconLocker.Lock()
	delete(nwState.BeaconState, peerPK)
	nwState.beaconLocker.Unlock()
	nwState.shardLocker.Lock()
	for cID := range nwState.ShardState {
		delete(nwState.ShardState[cID], peerPK)
	}
	nwState.shardLocker.Unlock()
}

type NetworkStateV2 struct {
	BeaconState       map[string]ChainState // map[<Committee Public Key>]Chainstate
	beaconLocker      *sync.RWMutex
	ShardState        map[byte]map[string]ChainState // map[<ShardID>]map[<Committee Public Key>]Chainstate
	shardLocker       *sync.RWMutex
	highwayIDOfPeerID *cache.Cache
}

func (nwState *NetworkStateV2) Init(numberOfShard int) {
	nwState.beaconLocker = new(sync.RWMutex)
	nwState.BeaconState = map[string]ChainState{}
	nwState.shardLocker = new(sync.RWMutex)
	nwState.ShardState = map[byte]map[string]ChainState{}
	for i := byte(0); i < byte(numberOfShard); i++ {
		nwState.ShardState[i] = map[string]ChainState{}
	}
	nwState.highwayIDOfPeerID = cache.New(common.MaxTimeKeepPeerState, common.MaxTimeKeepPeerState)
	nwState.highwayIDOfPeerID.OnEvicted(nwState.DeletePeerInfo)
}

func (nwState *NetworkStateV2) GetHWIDOfPeerID(
	pubKey string,
) (
	peer.ID,
	error,
) {
	hwIDFromCache, ok := nwState.highwayIDOfPeerID.Get(pubKey)
	if (!ok) || (hwIDFromCache == nil) {
		return "", errors.Errorf("Can not found highway ID for pubkey %v", pubKey)
	}
	return hwIDFromCache.(peer.ID), nil
}

func (nwState *NetworkStateV2) GetAllHWIDInfo() map[string]string {
	hwInfo := map[string]string{}
	hwInfoCached := nwState.highwayIDOfPeerID.Items()
	for pk, peerCached := range hwInfoCached {
		if peerCached.Object == nil {
			continue
		}
		pID := peerCached.Object.(peer.ID)
		hwInfo[pk] = pID.String()
	}
	return hwInfo
}

func (nwState *NetworkStateV2) SetHWIDOfPeerID(
	hwID peer.ID,
	pubKey string,
) error {
	nwState.highwayIDOfPeerID.Set(pubKey, hwID, common.MaxTimeKeepPeerState)
	return nil
}

func (nwState *NetworkStateV2) DeletePeerInfo(peerID string, highwayID interface{}) {
	logger.Infof("[delpeerstate] key %v", peerID)
	nwState.beaconLocker.Lock()
	delete(nwState.BeaconState, peerID)
	nwState.beaconLocker.Unlock()
	nwState.shardLocker.Lock()
	for cID := range nwState.ShardState {
		delete(nwState.ShardState[cID], peerID)
	}
	nwState.shardLocker.Unlock()
}
