package simulateutils

import "github.com/incognitochain/incognito-chain/wire"

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
	SupportShards []byte,
) ForkMaker {
	switch fType {
	case ONE_BLOCK:
		bftChs := map[int]chan wire.Message{}
		for _, cID := range SupportShards {
			bftChs[int(cID)] = make(chan wire.Message)
		}
		keyChs := map[int]chan interface{}{}
		for _, cID := range SupportShards {
			keyChs[int(cID)] = make(chan interface{})
		}
		return &OneBlockFork{
			SupportShard: SupportShards,
			CTable:       *NewCommitteeTable(),
			Scenes:       *NewScenario(),
			BFTsCh:       bftChs,
			KeysCh:       keyChs,
		}
	}
	return nil
}

type ForkMaker interface {
	BFTInbox(int) chan wire.Message
	KeyInbox(int) chan interface{}
	Start()
	Scenarioo
}

type Scenarioo interface {
	GetPubGroupWithMsg(wire.Message) map[string][]string
	SetScenes(SceneType, interface{})
	GetScenes(SceneType) interface{}
}

// type Scenee interface {
// 	IsTrigger(wire.Message) bool
// }
