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
	}
)
