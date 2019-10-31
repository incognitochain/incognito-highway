package process

import (
	"bytes"
	"encoding/json"
	fmt "fmt"
	"highway/common"
	logger "highway/customizelog"
	"io/ioutil"
	"os"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

type ChainData struct {
	ListMsgPeerStateOfShard   map[byte]CommitteeState //AllPeerState
	CurrentNetworkState       NetworkState
	CommitteePubkeyByPeerID   map[peer.ID]string
	PeerIDByCommitteePubkey   map[string]peer.ID
	ShardByCommitteePublicKey map[string]byte
	CommitteeKeyByMiningKey   map[string]string
	Locker                    *sync.RWMutex
}

func (chainData *ChainData) Init(
	filename string,
	numberOfShard,
	numberOfCandidate int,
) error {
	logger.Info("Init chaindata")
	chainData.ListMsgPeerStateOfShard = map[byte]CommitteeState{}
	chainData.Locker = &sync.RWMutex{}
	for i := 0; i < numberOfShard; i++ {
		chainData.ListMsgPeerStateOfShard[byte(i)] = map[string][]byte{}
	}
	chainData.CurrentNetworkState.Init(numberOfShard)
	chainData.CommitteePubkeyByPeerID = map[peer.ID]string{}
	chainData.PeerIDByCommitteePubkey = map[string]peer.ID{}
	chainData.ShardByCommitteePublicKey = map[string]byte{}
	chainData.CommitteeKeyByMiningKey = map[string]string{}
	err := chainData.InitGenesisCommitteeFromFile(filename, numberOfShard, numberOfCandidate)
	if err != nil {
		return err
	}
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
		err := validatorKey.FromString(validator)
		if err != nil {
			return 0, err
		}
		validatorMiningPK, err := validatorKey.MiningPublicKey()
		if err != nil {
			return 0, errors.New("Candidate " + validator + " not found 1")
		}

		if fullKey, ok := chainData.CommitteeKeyByMiningKey[validatorMiningPK]; ok {
			if fullKeyID, isExist := chainData.ShardByCommitteePublicKey[fullKey]; isExist {
				return fullKeyID, nil
			} else {
				logger.Infof("CommitteeKeyByMiningKey %v", chainData.ShardByCommitteePublicKey)
			}
		} else {
			logger.Infof("MiningKeyOfUser %v, len CommitteeKeyByMiningKey %v", validatorMiningPK, len(chainData.CommitteeKeyByMiningKey))
			i := 0
			for key, value := range chainData.CommitteeKeyByMiningKey {
				i++
				logger.Infof("First 4 candidate in keylist:\n MiningKey:%v \nCommitteeKey: %v", key, value)
				if i == 5 {
					break
				}
			}
		}
	}
	return 0, errors.New("Candidate " + validator + " not found 2")
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
				logger.Warnf("Committee publickey %v not found in PeerID map", committeePublicKey)
			}
		}

	}
	return nil, fmt.Errorf("Can not find any peer who has this block height: %v of committee %v", blkHeight, committeeID)
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
		err := validatorKey.FromString(validator)
		if err != nil {
			return nil, err
		}
		if len(validatorKey.IncPubKey) == 0 {
			return nil, errors.New("Peer ID for this candidate " + validator + " not found")
		}
		validatorMiningPK, err := validatorKey.MiningPublicKey()
		if err != nil {
			return nil, errors.New("Peer ID for this candidate " + validator + " not found")
		}
		if fullKey, ok := chainData.CommitteeKeyByMiningKey[validatorMiningPK]; ok {

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
	chainData.CommitteeKeyByMiningKey = map[string]string{}
	//#region Reading genesis committee key from keylist.json
	keyListFromFile := common.KeyList{}
	if filename != "" {
		jsonFile, err := os.Open(filename)
		if err != nil {
			logger.Error(err)
			return err
		}
		fmt.Printf("Successfully Opened %v\n", filename)
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal([]byte(byteValue), &keyListFromFile)
		if err != nil {
			return err
		}
	}

	for i := 0; i < numberOfCandidate; i++ {
		if i < len(keyListFromFile.Bc) {
			chainData.ShardByCommitteePublicKey[keyListFromFile.Bc[i].CommitteePubKey] = common.BEACONID
		}
	}
	logger.Infof("NumberOfShard %v, NumberOfCandidate %v", numberOfShard, numberOfCandidate)
	for j := 0; j < numberOfShard; j++ {
		for i := 0; i < numberOfCandidate; i++ {
			if i < len(keyListFromFile.Sh[j]) {
				chainData.ShardByCommitteePublicKey[keyListFromFile.Sh[j][i].CommitteePubKey] = byte(j)
			}
		}
	}
	//#endregion Reading genesis committee key from keylist.json
	logger.Infof("Result of reading from file:\nLen of keyListFromFile:\n\tBeacon %v\n\tShard: %v\nlen of ShardByCommittee %v", len(keyListFromFile.Bc), len(keyListFromFile.Sh[0]), len(chainData.ShardByCommitteePublicKey))
	for key := range chainData.ShardByCommitteePublicKey {
		committeePK := new(common.CommitteePublicKey)
		err := committeePK.FromString(key)
		if err != nil {
			logger.Info(err)
		} else {
			pkString, _ := committeePK.MiningPublicKey()
			logger.Info(pkString)
			chainData.CommitteeKeyByMiningKey[pkString] = key
		}
	}
	logger.Info("Result of init key %v", len(chainData.CommitteeKeyByMiningKey))
	return nil
}

func (chainData *ChainData) UpdateCommitteeState(
	committeeID byte,
	committeeState *CommitteeState,
) error {

	chainData.Locker.Lock()
	for key, peerState := range *committeeState {
		// logger.Infof(key)
		msgPeerState, err := ParsePeerStateData(string(peerState))
		if err != nil {
			logger.Error(errors.Wrapf(err, "Parse PeerState for committee %v false", committeeID))
			return err
		} else {
			// logger.Info(msgPeerState)
		}
		if committeeID == common.BEACONID {
			chainData.CurrentNetworkState.BeaconState[key] = newChainStateFromMsgPeerState(msgPeerState, committeeID)
		} else {
			chainData.CurrentNetworkState.ShardState[committeeID][key] = newChainStateFromMsgPeerState(msgPeerState, committeeID)
		}
	}
	defer chainData.Locker.Unlock()
	return nil
}

func (chainData *ChainData) UpdatePeerState(publisher string, data []byte) error {

	committeeID, err := chainData.GetCommitteeIDOfValidator(publisher)
	if err != nil {
		logger.Infof("This publisher not belong to current committee -%v- %v %v %v", publisher, common.BEACONID, common.NumberOfShard, committeeID)
		logger.Error(err)
		return err
	}
	chainData.Locker.Lock()
	if chainData.ListMsgPeerStateOfShard[byte(committeeID)] == nil {
		chainData.ListMsgPeerStateOfShard[byte(committeeID)] = map[string][]byte{}
	}
	if !bytes.Equal(chainData.ListMsgPeerStateOfShard[byte(committeeID)][publisher], data) {
		chainData.ListMsgPeerStateOfShard[byte(committeeID)][publisher] = data
	}
	chainData.Locker.Unlock()

	committeeState := chainData.ListMsgPeerStateOfShard[committeeID]
	err = chainData.UpdateCommitteeState(committeeID, &committeeState)
	if err != nil {
		logger.Error(err)
		return nil
	}
	// for committeeID, committeeState := range chainData.ListMsgPeerStateOfShard {
	// }
	return nil
}

func newChainStateFromMsgPeerState(
	msgPeerState *wire.MessagePeerState,
	committeeID byte,
	// candidateKey string,
) ChainState {
	var blkChainState blockchain.ChainState
	if committeeID == common.BEACONID {
		blkChainState = msgPeerState.Beacon
	} else {
		blkChainState = msgPeerState.Shards[committeeID]
	}
	return ChainState{
		Height:        blkChainState.Height,
		Timestamp:     blkChainState.Timestamp,
		BestStateHash: blkChainState.BestStateHash,
		BlockHash:     blkChainState.BlockHash,
	}
}
