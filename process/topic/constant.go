package topic

import "highway/common"

const (
	DoNothing byte = iota
	ProcessAndPublishAfter
	ProcessAndPublish
	WTFisThis
)

const (
	NODEPUB      = "-nodepub"
	NODESUB      = "-nodesub"
	NODESIDE     = "incnode"
	HIGHWAYSIDE  = "highway"
	noCIDInTopic = -1
)

const (
	CmdBlockShard         = "blockshard"
	CmdBlockBeacon        = "blockbeacon"
	CmdCrossShard         = "crossshard"
	CmdBlkShardToBeacon   = "blkshdtobcn"
	CmdTx                 = "tx"
	CmdCustomToken        = "txtoken"
	CmdPrivacyCustomToken = "txprivacytok"
	CmdPing               = "ping"

	CmdBFT       = "bft"
	CmdPeerState = "peerstate"
)

var (
	Message4Process = []string{
		"blockshard",
		"blockbeacon",
		"crossshard",
		"blkshdtobcn",
		"tx",
		"txtoken",
		"txprivacytok",
		"ping",
		"bft",
		"peerstate",
	}

	lvlAllowPubOfMsg = map[string]byte{
		"blockshard":   common.CANDIDATE,
		"blockbeacon":  common.CANDIDATE,
		"crossshard":   common.CANDIDATE,
		"blkshdtobcn":  common.CANDIDATE,
		"tx":           common.NORMAL,
		"txtoken":      common.NORMAL,
		"txprivacytok": common.NORMAL,
		"ping":         common.CANDIDATE,
		"bft":          common.CANDIDATE,
		"peerstate":    common.CANDIDATE,
	}
	lvlAllowSubOfMsg = map[string]byte{
		"blockshard":   common.NORMAL,
		"blockbeacon":  common.NORMAL,
		"crossshard":   common.CANDIDATE,
		"blkshdtobcn":  common.CANDIDATE,
		"tx":           common.NORMAL,
		"txtoken":      common.NORMAL,
		"txprivacytok": common.NORMAL,
		"ping":         common.CANDIDATE,
		"bft":          common.CANDIDATE,
		"peerstate":    common.PENDING,
	}
)
