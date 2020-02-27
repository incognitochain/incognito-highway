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

	DefaultRPCServerPort = 9330
	DefaultHWPeerID      = "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
)

var (
	ChainMaxCallRecvMsgSize     = 50 << 20 // 50 MBs per gRPC response
	ChainClientKeepaliveTime    = 10 * time.Minute
	ChainClientKeepaliveTimeout = 20 * time.Second
	ChainClientDialTimeout      = 5 * time.Second
	CacheNumCounters            = int64(100000)
	CacheMaxCost                = int64(8 * 2 << 30) // 8 GiB
	CacheBufferItems            = int64(64)

	MaxCallDepth         = int32(2)
	ChoosePeerBlockDelta = uint64(100)
	MaxBlocksPerRequest  = uint64(100)
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
)

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)
