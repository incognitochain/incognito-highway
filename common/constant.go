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
	NumberOfHighway      = 10

	GenesisBlockHeight = 1
)

var (
	ChainMaxCallRecvMsgSize     = 50 << 20 // 50 MBs per gRPC response
	ChainClientKeepaliveTime    = 10 * time.Minute
	ChainClientKeepaliveTimeout = 20 * time.Second
	ChainClientDialTimeout      = 5 * time.Second
	CacheNumCounters            = int64(100000)
	CacheMaxCost                = int64(2 * 1 << 30) // 2 GiB
	CacheBufferItems            = int64(64)

	MaxCallDepth         = int32(2)
	ChoosePeerBlockDelta = uint64(300)
	MaxBlocksPerRequest  = uint64(900)
	MaxTimePerRequest    = 30 * time.Second
	MaxTimeForSend       = 12 * time.Second

	TimeIntervalPublishStates = 5 * time.Second
	MaxTimeKeepPeerState      = 90 * time.Second
	MaxTimeKeepPubSubData     = 30 * time.Second

	RouteClientKeepaliveTime    = 20 * time.Minute
	RouteClientKeepaliveTimeout = 20 * time.Second
	RouteClientDialTimeout      = 5 * time.Second
	RouteKeepConnectionTimestep = 40 * time.Second
	RouteHighwayRemoveDeadline  = 60 * time.Minute
	MinStableDuration           = 2 * time.Minute

	BroadcastMsgEnlistTimestep = 1 * time.Minute

	PercentGetFromCache = 80
	TIMESLOT            = int64(10)
)

var TopicPrivate = map[string]struct{}{
	"bft": struct{}{},
}

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)

type ExpectedBlkByHeight struct {
	Height uint64
	Data   []byte
}

type ExpectedBlk struct {
	Height uint64
	Hash   []byte
	Data   []byte
}

type ExpectedBlkByHash struct {
	Hash []byte
	Data []byte
}
