package common

type CommitteePublicKey struct {
	IncPubKey    []byte
	MiningPubKey map[string][]byte
}

type MiningPublicKey map[string][]byte

var (
	SelfID string = "aaaaa"
	//Remove soon
	// CommitteeGenesis        map[string]byte
	// CommitteeKeyByMiningKey map[string]string
)

const (
	BEACONID      byte = 255
	NumberOfShard      = 2
	CommitteeSize      = 4
	BeaconRole         = "beacon"
	ShardRole          = "shard"

	CommitteeRole = "committee"
	PendingRole   = "pending"
	WaitingRole   = "waiting"
	NormalNode    = ""
)

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)
