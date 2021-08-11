package topic

import "highway/common"

const (
	NODEPUB      = "-nodepub"
	NODESUB      = "-nodesub"
	NODESIDE     = "incnode"
	HIGHWAYSIDE  = "highway"
	NoCIDInTopic = -2
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

	CmdFinishSync = "finishsync"
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
		"finishsync",
	}

	lvlAllowPubOfMsg = map[string]byte{
		"blockshard":   common.COMMITTEE,
		"blockbeacon":  common.COMMITTEE,
		"crossshard":   common.COMMITTEE,
		"blkshdtobcn":  common.COMMITTEE,
		"tx":           common.NORMAL,
		"txtoken":      common.NORMAL,
		"txprivacytok": common.NORMAL,
		"ping":         common.COMMITTEE,
		"bft":          common.COMMITTEE,
		"peerstate":    common.COMMITTEE,

		"finishsync": common.NORMAL,
	}
	lvlAllowSubOfMsg = map[string]byte{
		"blockshard":   common.NORMAL,
		"blockbeacon":  common.NORMAL,
		"crossshard":   common.COMMITTEE,
		"blkshdtobcn":  common.COMMITTEE,
		"tx":           common.NORMAL,
		"txtoken":      common.NORMAL,
		"txprivacytok": common.NORMAL,
		"ping":         common.COMMITTEE,
		"bft":          common.COMMITTEE,
		"peerstate":    common.NORMAL,

		"finishsync": common.COMMITTEE,
	}
)
