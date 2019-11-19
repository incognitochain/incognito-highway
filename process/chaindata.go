package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"highway/common"
	"highway/proto"
	"io/ioutil"
	"os"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
	ic "github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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
	masternode                peer.ID
}

type PeerWithBlk struct {
	ID     peer.ID
	Height uint64
}

func (chainData *ChainData) Init(
	filename string,
	numberOfShard,
	numberOfCandidate int,
	masternode peer.ID,
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
	chainData.masternode = masternode
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
			// logger.Infof("MiningKeyOfUser %v, len CommitteeKeyByMiningKey %v", validatorMiningPK, len(chainData.CommitteeKeyByMiningKey))
			// i := 0
			// for key, value := range chainData.CommitteeKeyByMiningKey {
			// 	i++
			// 	logger.Debugf("First 4 candidate in keylist:\n MiningKey:%v \nCommitteeKey: %v", key, value)
			// 	if i == 5 {
			// 		break
			// 	}
			// }
		}
	}
	return 0, errors.New("Candidate " + validator + " not found 2")
}

func (chainData *ChainData) GetPeerHasBlk(
	blkHeight uint64,
	committeeID byte,
) (
	[]PeerWithBlk,
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
			return nil, errors.New("committeeID " + string(committeeID) + " not found")
		}
	}
	peers := []PeerWithBlk{}
	for committeePublicKey, nodeState := range committeeState {
		if peerID, ok := chainData.PeerIDByCommitteePubkey[committeePublicKey]; ok {
			peers = append(peers, PeerWithBlk{
				ID:     peerID,
				Height: nodeState.Height,
			})
		} else {
			logger.Warnf("Committee publickey %v not found in PeerID map", committeePublicKey)
		}
	}

	// Sort based on block height
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Height > peers[j].Height
	})
	return peers, nil
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
	logger.Infof("NumberOfShard %v, NumberOfCandidate %v", numberOfShard, numberOfCandidate)

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
	//#endregion Reading genesis committee key from keylist.json

	// Cut off keylist.json and update
	keyListFromFile.Bc = keyListFromFile.Bc[:]
	for k := range keyListFromFile.Sh {
		if k >= numberOfShard {
			delete(keyListFromFile.Sh, k)
		}
	}
	for j := range keyListFromFile.Sh {
		keyListFromFile.Sh[j] = keyListFromFile.Sh[j][:numberOfCandidate]
	}
	chainData.updateCommitteePublicKey(&keyListFromFile)

	logger.Infof("Result of reading from file:\nLen of keyListFromFile:\n\tBeacon %v\n\tShard: %v\nlen of ShardByCommittee %v", len(keyListFromFile.Bc), len(keyListFromFile.Sh[0]), len(chainData.ShardByCommitteePublicKey))
	logger.Infof("Result of init key %v", len(chainData.CommitteeKeyByMiningKey))
	return nil
}

// updateCommitteePublicKey saves the publickeys of all validators and update
// the mapping from miningkey to publickey
func (chainData *ChainData) updateCommitteePublicKey(keys *common.KeyList) {
	chainData.Locker.Lock()
	defer chainData.Locker.Unlock()
	for _, val := range keys.Bc {
		chainData.ShardByCommitteePublicKey[val.CommitteePubKey] = common.BEACONID
	}
	for j, vals := range keys.Sh {
		for _, val := range vals {
			chainData.ShardByCommitteePublicKey[val.CommitteePubKey] = byte(j)
		}
	}
	for key := range chainData.ShardByCommitteePublicKey {
		committeePK := new(common.CommitteePublicKey)
		err := committeePK.FromString(key)
		if err != nil {
			logger.Error(err)
		} else {
			pkString, _ := committeePK.MiningPublicKey()
			// logger.Debug(pkString)
			chainData.CommitteeKeyByMiningKey[pkString] = key
		}
	}
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

// ProcessChainCommittee receives all messages containing the new
// committee published by masternode and update the list of committee members
func (chainData *ChainData) ProcessChainCommitteeMsg(sub *pubsub.Subscription) {

	//Message just contains CommitteeKey, Pending, so, how about waiting?

	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		logger.Info("Received new committee")
		if err != nil {
			logger.Error(err)
			continue
		}

		// TODO(@0xbunyip): check if msg.From can be maniconfiguration: driver=usbhid maxpower=200mA speed=12Mbit/s
		if peer.ID(msg.From) != chainData.masternode {
			from := peer.IDB58Encode(peer.ID(msg.From))
			exp := peer.IDB58Encode(chainData.masternode)
			logger.Warnf("Received NewCommittee from unauthorized source, expect from %+v, got from %+v, data %+v", from, exp, msg.Data)
			continue
		}

		logger.Info("Saving new committee")
		comm := &incognitokey.ChainCommittee{}
		if err := json.Unmarshal(msg.Data, comm); err != nil {
			logger.Error(err)
			continue
		}

		// Update chain committee
		keys, err := getKeyListFromMessage(comm)
		if err != nil {
			logger.Error(err)
			continue
		}
		chainData.updateCommitteePublicKey(keys)
	}
}

func getKeyListFromMessage(comm *incognitokey.ChainCommittee) (*common.KeyList, error) {
	// TODO(@0xbunyip): handle epoch
	keys := &common.KeyList{Sh: map[int][]common.Key{}}
	for _, k := range comm.BeaconCommittee {
		cpk, err := k.ToBase58()
		if err != nil {
			return nil, errors.Wrapf(err, "key: %+v", k)
		}
		keys.Bc = append(keys.Bc, common.Key{CommitteePubKey: cpk})
	}

	for s, vals := range comm.AllShardCommittee {
		for _, val := range vals {
			cpk, err := val.ToBase58()
			if err != nil {
				return nil, errors.Wrapf(err, "key: %+v", val)
			}
			keys.Sh[int(s)] = append(keys.Sh[int(s)], common.Key{CommitteePubKey: cpk})
		}
	}
	return keys, nil
}

func GetUserRole(cid int) *proto.UserRole {
	// TODO(@0xbunyip): support pending, waiting and normal role
	layer := ""
	role := ""
	shard := int(common.BEACONID)
	if cid == int(common.BEACONID) {
		layer = ic.BeaconRole
		role = ic.CommitteeRole
		shard = -1
	} else if cid != -1 { // other than NORMAL
		layer = ic.ShardRole
		role = ic.CommitteeRole
		shard = cid
	} else {
		layer = ""
		role = ""
		shard = -2
	}
	return &proto.UserRole{
		Layer: layer,
		Role:  role,
		Shard: int32(shard),
	}
}

// GetCommitteeInfoOfPublicKey With input public key, return role of it, if the role is COMMITTEE, return committee ID of it.
// ROLE value is defined in common/constant.go
// TODO Add pending list into Chaindata, update this function
func (chainData *ChainData) GetCommitteeInfoOfPublicKey(
	pk string,
) (
	byte,
	int,
) {
	cID, err := chainData.GetCommitteeIDOfValidator(pk)
	if err != nil {
		return common.NORMAL, -1
	}
	return common.COMMITTEE, int(cID)
}
