package common

import "time"

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

	MaxCallDepth         = 2
	ChoosePeerBlockDelta = 100
	MaxBlocksPerRequest  = 100
	MaxTimePerRequest    = 20 * time.Second

	TimeIntervalPublishStates = 5 * time.Second
	MaxTimeKeepPeerState      = 90 * time.Second
	MaxTimeKeepPubSubData     = 30 * time.Second

	RouteClientKeepaliveTime    = 20 * time.Minute
	RouteClientKeepaliveTimeout = 20 * time.Second
	RouteClientDialTimeout      = 5 * time.Second
	RouteKeepConnectionTimestep = 40 * time.Second
	RouteHighwayKeepaliveTime   = 40 * time.Second

	BroadcastMsgEnlistTimestep = 1 * time.Minute

	DefaultRPCServerPort = 9330
	DefaultHWPeerID      = "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
)

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)
