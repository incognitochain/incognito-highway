package common

var (
	SelfID string = "aaaaa"
	//Remove soon
	// CommitteeGenesis        map[string]byte
	// CommitteeKeyByMiningKey map[string]string
)

const (
	BEACONID      byte = 255
	NumberOfShard      = 8
	CommitteeSize      = 32
	BeaconRole         = "beacon"
	ShardRole          = "shard"

	CommitteeRole = "committee"
	PendingRole   = "pending"
	WaitingRole   = "waiting"
	NormalNode    = ""

	MaxBlocksPerRequest = 100
)

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)
