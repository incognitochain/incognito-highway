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

	MaxBlocksPerRequest         = 100
	ChoosePeerBlockDelta        = 100
	RouteClientKeepaliveTime    = 20 * time.Minute
	RouteClientKeepaliveTimeout = 20 * time.Second

	DefaultRPCServerPort = 9330
	DefaultHWPeerID      = "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
)

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)
