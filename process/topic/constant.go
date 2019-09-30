package topic

const (
	MessBFT byte = iota
	//
	MessGetBlockBeacon
	MessGetBlockShard
	MessGetCrossShard
	MessGetShardToBeacon
	MessBlockShard
	MessBlockBeacon
	MessCrossShard
	MessBlkShardToBeacon
	MessGetAddr
	MessAddr
	MessPing
	MessPeerState
)

const (
	CmdGetBlockBeacon     = "getblkbeacon"
	CmdGetBlockShard      = "getblkshard"
	CmdGetCrossShard      = "getcrossshd"
	CmdGetShardToBeacon   = "getshdtobcn"
	CmdBlockShard         = "blockshard"
	CmdBlockBeacon        = "blockbeacon"
	CmdCrossShard         = "crossshard"
	CmdBlkShardToBeacon   = "blkshdtobcn"
	CmdTx                 = "tx"
	CmdCustomToken        = "txtoken"
	CmdPrivacyCustomToken = "txprivacytok"
	CmdVersion            = "version"
	CmdVerack             = "verack"
	CmdGetAddr            = "getaddr"
	CmdAddr               = "addr"
	CmdPing               = "ping"

	// POS Cmd
	CmdBFT       = "bft"
	CmdPeerState = "peerstate"

	// heavy message check cmd
	CmdMsgCheck     = "msgcheck"
	CmdMsgCheckResp = "msgcheckresp"
)

var Message4Process = []string{
	"a",
	"getblkbeacon",
	"getblkshard",
	"getcrossshd",
	"getshdtobcn",
	"blockshard",
	"blockbeacon",
	"crossshard",
	"blkshdtobcn",
	"tx",
	"txtoken",
	"txprivacytok",
	"version",
	"verack",
	"getaddr",
	"addr",
	"ping",
	"bft",
	"peerstate",
}
