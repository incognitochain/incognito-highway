package simulateutils

import (
	libp2p "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/incognitochain/incognito-chain/consensus/blsbftv2"
	"github.com/incognitochain/incognito-chain/wire"
)

type ForkType int
type SceneType int

const (
	ONE_BLOCK ForkType = 0
)

const (
	TRIGGER_BY_TIMESLOT    SceneType = 0
	TRIGGER_BY_BLOCKHEIGHT SceneType = 1
	TRIGGER_BY_TXID        SceneType = 2
)

func NewForkMaker(
	fType ForkType,
	committeeInfo *CommitteeTable,
	SupportShards []byte,
	pubSub *libp2p.PubSub,
) ForkMaker {
	switch fType {
	case ONE_BLOCK:
		pbftChs := map[int]chan MsgBFT{}
		for _, cID := range SupportShards {
			pbftChs[int(cID)] = make(chan MsgBFT, 1000)
		}
		vbftChs := map[int]chan MsgBFT{}
		for _, cID := range SupportShards {
			vbftChs[int(cID)] = make(chan MsgBFT, 1000)
		}
		keyChs := map[int]chan interface{}{}
		for _, cID := range SupportShards {
			keyChs[int(cID)] = make(chan interface{})
		}
		return &OneBlockFork{
			SupportShard:  SupportShards,
			CTable:        *NewCommitteeTable(),
			Scenes:        *NewScenarioBasic(SupportShards),
			BS:            NewTriggerByBlock(),
			PBFTsCh:       pbftChs,
			VBFTsCh:       vbftChs,
			KeysCh:        keyChs,
			CommitteeInfo: committeeInfo,
			PubSub:        pubSub,
		}
	}
	return nil
}

type MsgPropose struct {
	MsgPropose *blsbftv2.BFTPropose
	BlkHeight  uint64
	BlkHash    string
	ShardID    int
}

type BlkInfoByHash struct {
	BlkHeight uint64
	Round     int
	ShardID   int
}

type MsgVote struct {
	MsgVote *blsbftv2.BFTVote
	BlkHash string
	ShardID int
}

type BlkInfoByHeight struct {
	BlkHash string
	Round   int
	ShardID int
	Propose MsgBFT
	Vote    []wire.Message
}

type MsgBFT struct {
	ProposeIndex int
	Msg          wire.Message
	Topic        string
}

type PPublishInfo struct {
	Height        uint64
	StartTimeSlot int64
	MsgBFTByRound map[int]MsgBFT
	ForkScene     *BasicForkScene
}

type VPublishInfo struct {
	Height        int
	MsgBFTByRound map[int]MsgBFT
	ForkScene     *BasicForkScene
}

type ForkMaker interface {
	PBFTInbox(int) chan MsgBFT
	VBFTInbox(int) chan MsgBFT
	KeyInbox(int) chan interface{}
	Start()
	Scenarioo
	IsTrigger(wire.Message, int) bool
}

type Scenarioo interface {
	// GetPubGroupWithMsg(wire.Message) map[string][]string
	SetScenes(SceneType, []byte, int) error
	// GetScenes(SceneType) (interface{})
}

type Scenee interface {
	IsTrigger(wire.Message) (bool, BasicForkScene)
}
