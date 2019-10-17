package process

import (
	"bytes"
	"encoding/json"
	"errors"
	fmt "fmt"
	"highway/common"
	logger "highway/customizelog"
	"io/ioutil"
	"os"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
)

type ChainData struct {
	ListMsgPeerStateOfShard   map[byte]CommitteeState //AllPeerState
	CurrentNetworkState       NetworkState
	CommitteePubkeyByPeerID   map[peer.ID]string
	PeerIDByCommitteePubkey   map[string]peer.ID
	ShardByCommitteePublicKey map[string]byte
	MiningKeyByCommitteeKey   map[string]string
	Locker                    sync.RWMutex
}

func (chainData *ChainData) Init(
	filename string,
	numberOfShard,
	numberOfCandidate int,
) error {
	chainData.InitGenesisCommitteeFromFile(filename, numberOfShard, numberOfCandidate)
	chainData.ListMsgPeerStateOfShard = map[byte]CommitteeState{}
	for i := 0; i < numberOfShard; i++ {
		chainData.ListMsgPeerStateOfShard[byte(i)] = map[string][]byte{}
	}
	for i := 0; i < numberOfShard; i++ {
		chainData.ListMsgPeerStateOfShard[byte(i)] = map[string][]byte{}
	}
	chainData.CurrentNetworkState.Init(numberOfShard)
	chainData.CommitteePubkeyByPeerID = map[peer.ID]string{}
	chainData.PeerIDByCommitteePubkey = map[string]peer.ID{}
	chainData.ShardByCommitteePublicKey = map[string]byte{}
	chainData.MiningKeyByCommitteeKey = map[string]string{}
	return nil
}

func (chainData *ChainData) GetCommitteeIDOfValidator(
	validator string,
) (
	byte,
	error,
) {
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if id, exist := chainData.ShardByCommitteePublicKey[validator]; exist {
		return id, nil
	} else {
		validatorKey := new(common.CommitteePublicKey)
		validatorKey.FromString(validator)
		validatorMiningPK, err := validatorKey.MiningPublicKey()
		if err != nil {
			return 0, errors.New("Candidate " + validator + " not found")
		}
		if fullKey, ok := common.MiningKeyByCommitteeKey[validatorMiningPK]; ok {
			if fullKeyID, isExist := chainData.ShardByCommitteePublicKey[fullKey]; isExist {
				return fullKeyID, nil
			}
		}
	}
	return 0, errors.New("Candidate " + validator + " not found")
}

func (chainData *ChainData) GetPeerHasBlk(
	blkHeight uint64,
	committeeID byte,
) (
	*peer.ID,
	error,
) {
	var exist bool
	var committeeState map[string]ChainState
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if committeeID == common.BEACONID {
		committeeState = chainData.CurrentNetworkState.BeaconState
	} else {
		if committeeState, exist = chainData.CurrentNetworkState.ShardState[committeeID]; !exist {
			return nil, errors.New("CommitteeID " + string(committeeID) + " not found")
		}
	}
	for committeePublicKey, nodeState := range committeeState {
		//TODO get random
		if nodeState.Height > blkHeight {
			if peerID, ok := chainData.PeerIDByCommitteePubkey[committeePublicKey]; ok {
				return &peerID, nil
			} else {
				return nil, errors.New("Committee publickey not found in PeerID map")
			}
		}

	}
	return nil, errors.New(fmt.Sprintf("Can not find any peer who has this block height: %v of committee %v", blkHeight, committeeID))
}

func (chainData *ChainData) GetPeerIDOfValidator(
	validator string,
) (
	*peer.ID,
	error,
) {
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if peerID, exist := chainData.PeerIDByCommitteePubkey[validator]; exist {
		return &peerID, nil
	} else {
		validatorKey := new(common.CommitteePublicKey)
		validatorKey.FromString(validator)
		if len(validatorKey.IncPubKey) == 0 {
			return nil, errors.New("Peer ID for this candidate " + validator + " not found")
		}
		validatorMiningPK, err := validatorKey.MiningPublicKey()
		if err != nil {
			return nil, errors.New("Peer ID for this candidate " + validator + " not found")
		}
		if fullKey, ok := common.MiningKeyByCommitteeKey[validatorMiningPK]; ok {

			if peerID, exist := chainData.PeerIDByCommitteePubkey[fullKey]; exist {
				return &peerID, nil
			}
			return nil, errors.New("Peer ID for this candidate " + validator + " not found")
		}
	}
	return nil, errors.New("Candidate " + validator + " not found")
}

func (chainData *ChainData) UpdatePeerIDOfCommitteePubkey(
	candidate string,
	peerID *peer.ID,
) {
	chainData.Locker.Lock()
	chainData.CommitteePubkeyByPeerID[*peerID] = candidate
	chainData.PeerIDByCommitteePubkey[candidate] = *peerID
	chainData.Locker.Unlock()
}

func (chainData *ChainData) InitGenesisCommitteeFromFile(
	filename string,
	numberOfShard,
	numberOfCandidate int,
) error {
	chainData.Locker.Lock()
	defer chainData.Locker.Unlock()
	chainData.ShardByCommitteePublicKey = map[string]byte{}
	chainData.MiningKeyByCommitteeKey = map[string]string{}
	keyListFromFile := common.KeyList{}
	if filename != "" {
		jsonFile, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("Successfully Opened %v\n", filename)
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal([]byte(byteValue), &keyListFromFile)

	}

	for i := 0; i < numberOfCandidate; i++ {
		if i < len(keyListFromFile.Bc) {
			chainData.ShardByCommitteePublicKey[keyListFromFile.Bc[i].CommitteePubKey] = common.BEACONID
		}
	}
	for j := 0; j < numberOfShard; j++ {
		for i := 0; i < numberOfCandidate; i++ {
			if i < len(keyListFromFile.Sh[j]) {
				chainData.ShardByCommitteePublicKey[keyListFromFile.Sh[j][i].CommitteePubKey] = byte(j)
			}
		}
	}
	for key, _ := range chainData.ShardByCommitteePublicKey {
		committeePK := new(common.CommitteePublicKey)
		err := committeePK.FromString(key)
		if err != nil {
			logger.Info(err)
		} else {
			pkString, _ := committeePK.MiningPublicKey()
			chainData.MiningKeyByCommitteeKey[pkString] = key
		}
	}
	return nil
}

func (chainData *ChainData) UpdateCommitteeState(
	committeeID byte,
	committeeState *CommitteeState,
) error {
	chainData.Locker.Lock()
	defer chainData.Locker.Unlock()
	return nil
}

func (chainData *ChainData) UpdatePeerState(publisher string, data []byte) error {

	committeeID, err := chainData.GetCommitteeIDOfValidator(publisher)
	if err != nil {
		logger.Infof("This publisher not belong to current committee -%v- %v %v %v", publisher, common.BEACONID, common.NumberOfShard, committeeID)
		return err
	}
	chainData.Locker.Lock()
	if !bytes.Equal(chainData.ListMsgPeerStateOfShard[byte(committeeID)][publisher], data) {
		chainData.ListMsgPeerStateOfShard[byte(committeeID)][publisher] = data
	}
	chainData.Locker.Unlock()

	for committeeID, committeeState := range chainData.ListMsgPeerStateOfShard {
		err := chainData.UpdateCommitteeState(committeeID, &committeeState)
		if err != nil {
			logger.Error(err)
			return nil
		}
	}
	return nil
}