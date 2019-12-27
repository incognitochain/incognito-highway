package common

import "time"

var (
	SelfID string = "aaaaa"
	//Remove soon
	// CommitteeGenesis        map[string]byte
	// CommitteeKeyByMiningKey map[string]string
)

const (
	BEACONID      byte = 255
	NumberOfShard      = 8
	BeaconRole         = "beacon"
	ShardRole          = "shard"

	CommitteeRole = "committee"
	PendingRole   = "pending"
	WaitingRole   = "waiting"
	NormalNode    = ""

	ChainClientKeepaliveTime    = 10 * time.Minute
	ChainClientKeepaliveTimeout = 20 * time.Second
	ChainClientDialTimeout      = 5 * time.Second
	MaxBlocksPerRequest         = 100

	RouteClientKeepaliveTime    = 10 * time.Minute
	RouteClientKeepaliveTimeout = 20 * time.Second
	RouteClientDialTimeout      = 5 * time.Second

	DefaultRPCServerPort = 9330
	DefaultHWPeerID      = "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
)

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)
