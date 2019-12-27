package chaindata

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

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
	HighwayIDOfPublicKey map[string]string
	hwInfoLocker         *sync.RWMutex
	PeerStateLastUpdate  map[string]time.Time
	peerInfoLocker       *sync.RWMutex
	MaxTimeKeepPeerState uint64
}

func (nwState *NetworkState) Init(numberOfShard int) {
	nwState.beaconLocker = new(sync.RWMutex)
	nwState.BeaconState = map[string]ChainState{}
	nwState.shardLocker = new(sync.RWMutex)
	nwState.ShardState = map[byte]map[string]ChainState{}
	for i := byte(0); i < byte(numberOfShard); i++ {
		nwState.ShardState[i] = map[string]ChainState{}
	}
	nwState.HighwayIDOfPublicKey = map[string]string{}
	nwState.hwInfoLocker = new(sync.RWMutex)
	nwState.PeerStateLastUpdate = map[string]time.Time{}
	nwState.peerInfoLocker = new(sync.RWMutex)
}

func (nwState *NetworkState) GetHWIDOfPubKey(
	pubKey string,
) (
	string,
	error,
) {
	nwState.hwInfoLocker.RLock()
	defer nwState.hwInfoLocker.RUnlock()
	if hwID, ok := nwState.HighwayIDOfPublicKey[pubKey]; ok {
		return hwID, nil
	}
	return "", errors.Errorf("Can not found highway ID for pubkey %v", pubKey)
}

func (nwState *NetworkState) SetHWIDOfPubKey(
	hwID string,
	pubKey string,
) error {
	nwState.hwInfoLocker.Lock()
	defer nwState.hwInfoLocker.Unlock()
	nwState.HighwayIDOfPublicKey[pubKey] = hwID
	return nil
}

func (nwState *NetworkState) GetLastUpdateOfPubKey(
	pubKey string,
) (
	time.Time,
	error,
) {
	nwState.peerInfoLocker.RLock()
	defer nwState.peerInfoLocker.RUnlock()
	if lastTime, ok := nwState.PeerStateLastUpdate[pubKey]; ok {
		return lastTime, nil
	}

	return time.Time{}, errors.Errorf("Can not found last time updated of pubkey %v", pubKey)
}

func (nwState *NetworkState) SetLastUpdateOfPubKey(
	pubKey string,
	lastTime time.Time,
) error {
	nwState.peerInfoLocker.Lock()
	defer nwState.peerInfoLocker.Unlock()
	nwState.PeerStateLastUpdate[pubKey] = lastTime
	return nil
}

// TODO Complete in next pull request
// func (nwState *NetworkState) DeleteOutdatedPeerInfo() []string {
// 	listOutdated := []string{}
// 	currentTime := time.Now()
// 	nwState.peerInfoLocker.Lock()
// 	defer nwState.peerInfoLocker.Unlock()
// 	for peerPK, lastTime := range nwState.PeerStateLastUpdate {
// 		if uint64(currentTime.Sub(lastTime).Milliseconds()) >= nwState.MaxTimeKeepPeerState {
// 			delete(nwState.PeerStateLastUpdate, peerPK)
// 			delete(nwState.HighwayIDOfPublicKey, peerPK)
// 			nwState.beaconLocker.Lock()
// 			delete(nwState.BeaconState, peerPK)
// 			nwState.beaconLocker.Unlock()
// 			nwState.shardLocker.Lock()
// 			for _, shardState :=
// 			nwState.shardLocker.Unlock()
// 		}
// 	}

// 	return []string{}
// }
