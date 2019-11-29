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
	ListMsgPeerStateOfShard    map[byte]CommitteeState //AllPeerState
	CurrentNetworkState        NetworkState
	MiningPubkeyByPeerID       map[peer.ID]string
	PeerIDByMiningPubkey       map[string]peer.ID
	ShardByMiningPubkey        map[string]byte
	ShardPendingByMiningPubkey map[string]byte
	Locker                     *sync.RWMutex
	masternode                 peer.ID
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
	chainData.MiningPubkeyByPeerID = map[peer.ID]string{}
	chainData.PeerIDByMiningPubkey = map[string]peer.ID{}
	chainData.ShardPendingByMiningPubkey = map[string]byte{}
	chainData.ShardByMiningPubkey = map[string]byte{}
	chainData.masternode = masternode
	err := chainData.InitGenesisCommitteeFromFile(filename, numberOfShard, numberOfCandidate)
	if err != nil {
		return err
	}
	return nil
}

func (chainData *ChainData) GetCommitteeIDOfValidator(validator string) (byte, error) {
	miningPubkey, err := extractMiningKey(validator)
	if err != nil {
		return 0, err
	}

	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if cid, ok := chainData.ShardByMiningPubkey[miningPubkey]; ok {
		return cid, nil
	}
	return 0, errors.New("candidate " + validator + " not found 2")
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
	for miningPubkey, nodeState := range committeeState {
		if peerID, ok := chainData.PeerIDByMiningPubkey[miningPubkey]; ok {
			peers = append(peers, PeerWithBlk{
				ID:     peerID,
				Height: nodeState.Height,
			})
		} else {
			logger.Warnf("Committee publickey %v not found in PeerID map", miningPubkey)
		}
	}

	// Sort based on block height
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Height > peers[j].Height
	})
	return peers, nil
}

func (chainData *ChainData) GetPeerIDOfValidator(validator string) (*peer.ID, error) {
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	miningPubkey, err := extractMiningKey(validator)
	if err != nil {
		return nil, err
	}

	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if pid, ok := chainData.PeerIDByMiningPubkey[miningPubkey]; ok {
		return &pid, nil
	}
	return nil, errors.New("candidate " + validator + " not found")
}

// UpdateCommittee saves peerID, mining pubkey and committeeID of a validator
func (chainData *ChainData) UpdateCommittee(pubkey string, peerID peer.ID, cid byte) error {
	// Convert from CommitteePubkey to MiningPubKey if user submitted one
	miningPubkey, err := extractMiningKey(pubkey)
	if err != nil {
		return err
	}

	// Map between mining pubkey and peerID
	chainData.Locker.Lock()
	defer chainData.Locker.Unlock()
	chainData.MiningPubkeyByPeerID[peerID] = miningPubkey
	chainData.PeerIDByMiningPubkey[miningPubkey] = peerID
	chainData.ShardByMiningPubkey[miningPubkey] = cid
	return nil
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
	for k := range keyListFromFile.Sh {
		if k >= numberOfShard {
			delete(keyListFromFile.Sh, k)
		}
	}
	for j := range keyListFromFile.Sh {
		keyListFromFile.Sh[j] = keyListFromFile.Sh[j][:numberOfCandidate]
	}
	chainData.updateMiningKeys(&keyListFromFile)

	logger.Infof("Result of reading from file:\nLen of keyListFromFile:\n\tBeacon %v\n\tShard: %v\nlen of ShardByCommittee %v", len(keyListFromFile.Bc), len(keyListFromFile.Sh[0]), len(chainData.ShardByMiningPubkey))
	return nil
}

// updateMiningKeys saves the publickeys of all validators
func (chainData *ChainData) updateMiningKeys(keys *common.KeyList) error {
	chainData.Locker.Lock()
	defer chainData.Locker.Unlock()
	for _, val := range keys.Bc {
		miningPubkey, err := extractMiningKey(val.CommitteePubKey)
		if err != nil {
			return err
		}
		chainData.ShardByMiningPubkey[miningPubkey] = common.BEACONID
	}
	for j, vals := range keys.Sh {
		for _, val := range vals {
			miningPubkey, err := extractMiningKey(val.CommitteePubKey)
			if err != nil {
				return err
			}
			chainData.ShardByMiningPubkey[miningPubkey] = byte(j)
		}
	}
	for j, pends := range keys.ShPend {
		for _, pend := range pends {
			miningPubkey, err := extractMiningKey(pend.CommitteePubKey)
			if err != nil {
				return err
			}
			chainData.ShardPendingByMiningPubkey[miningPubkey] = byte(j)
		}
	}
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
		logger.Infof("This publisher not belong to current committee %v %v", publisher, committeeID)
		return err
	}

	// Save peerstate by miningPubkey
	miningPubkey, err := extractMiningKey(publisher)
	if err != nil {
		return err
	}

	chainData.Locker.Lock()
	if chainData.ListMsgPeerStateOfShard[byte(committeeID)] == nil {
		chainData.ListMsgPeerStateOfShard[byte(committeeID)] = map[string][]byte{}
	}
	if !bytes.Equal(chainData.ListMsgPeerStateOfShard[byte(committeeID)][miningPubkey], data) {
		chainData.ListMsgPeerStateOfShard[byte(committeeID)][miningPubkey] = data
	}
	chainData.Locker.Unlock()

	committeeState := chainData.ListMsgPeerStateOfShard[committeeID]
	return chainData.UpdateCommitteeState(committeeID, &committeeState)
}

func (chainData *ChainData) GetMiningPubkeyFromPeerID(pid peer.ID) string {
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	return chainData.MiningPubkeyByPeerID[pid]
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
		chainData.updateMiningKeys(keys)
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

	// Shard's pending validators
	for s, pends := range comm.AllShardPending {
		for _, pend := range pends {
			cpk, err := pend.ToBase58()
			if err != nil {
				return nil, errors.Wrapf(err, "key: %+v", pend)
			}
			keys.ShPend[int(s)] = append(keys.ShPend[int(s)], common.Key{CommitteePubKey: cpk})
		}
	}

	return keys, nil
}

func GetUserRole(role string, cid int) *proto.UserRole {
	layer := ""
	if cid == int(common.BEACONID) {
		layer = ic.BeaconRole
	} else if cid != -1 { // other than NORMAL
		layer = ic.ShardRole
	} else {
		layer = ""
		role = ""
	}
	return &proto.UserRole{
		Layer: layer,
		Role:  role,
		Shard: int32(cid),
	}
}

// extractMiningKey receives either CommitteePubKey or MiningPubKey;
// for CommitteePubKey, it converts to MiningPubKey and returns;
// otherwise, keep the key as it is
func extractMiningKey(pubkey string) (string, error) {
	key := new(common.CommitteePublicKey)
	if err := key.FromString(pubkey); err != nil {
		return "", errors.WithMessagef(err, "pubkey: %s", pubkey)
	}

	key.IncPubKey = nil
	miningPubkey, err := key.ToBase58()
	if err != nil {
		return "", errors.WithMessagef(err, "pubkey: %s", pubkey)
	}
	return miningPubkey, nil
}
