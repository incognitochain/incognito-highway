package process

type ChainState struct {
	Timestamp     int64
	Height        uint64
	BlockHash     [32]byte
	BestStateHash [32]byte
}

type CommitteeState map[string][]byte

// var AllPeerState map[byte]CommitteeState
// var currentAllPeerStateData []byte

// NetworkState contains all of chainstate of node, which in committee and connected to proxy
type NetworkState struct {
	BeaconState map[string]ChainState          // map[<Committee Public Key>]Chainstate
	ShardState  map[byte]map[string]ChainState // map[<ShardID>]map[<Committee Public Key>]Chainstate
}

var NWState NetworkState

func (nwState *NetworkState) Init(numberOfShard int) {
	nwState.BeaconState = map[string]ChainState{}
	nwState.ShardState = map[byte]map[string]ChainState{}
	for i := byte(0); i < byte(numberOfShard); i++ {
		nwState.ShardState[i] = map[string]ChainState{}
	}
}
