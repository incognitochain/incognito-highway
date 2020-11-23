package simulateutils

import (
	"encoding/json"
	"fmt"
	"highway/common"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/consensus/blsbftv2"
	"github.com/incognitochain/incognito-chain/wire"
)

type OneBlockFork struct {
	CTable       CommitteeTable
	Scenes       Scenario
	BFTsCh       map[int]chan wire.Message
	KeysCh       map[int]chan interface{}
	SupportShard []byte
	status       string
}

func (f *OneBlockFork) BFTInbox(cID int) chan wire.Message {
	return f.BFTsCh[cID]
}

func (f *OneBlockFork) KeyInbox(cID int) chan interface{} {
	return f.KeysCh[cID]
}

func (f *OneBlockFork) SetScenes(SceneType, interface{}) {
	return
}

func (f *OneBlockFork) GetScenes(SceneType) interface{} {
	return nil
}

func (f *OneBlockFork) GetPubGroupWithMsg(wire.Message) map[string][]string {
	return nil
}

// func

func (f *OneBlockFork) Start() {
	f.status = "running"
	for _, cID := range f.SupportShard {
		go f.WatchingChain(int(cID))
	}
	go f.CountingTimeSlot()
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

type BFTMoreInfo struct {
	MsgPropose *blsbftv2.BFTPropose
	MsgVote    *blsbftv2.BFTVote
	BlkHeight  uint64
	ShardID    int
}

func ParseBFTPropose(x *wire.MessageBFT) MsgPropose {
	var msgPropose blsbftv2.BFTPropose
	err := json.Unmarshal(x.Content, &msgPropose)
	if err != nil {
		fmt.Errorf("%v", err)
		return MsgPropose{}
	}
	res := MsgPropose{}
	res.MsgPropose = &msgPropose
	msgPropose.PeerID = x.PeerID
	if x.ChainKey != "beacon" {
		y := blockchain.NewShardBlock()
		y.UnmarshalJSON(msgPropose.Block)
		res.BlkHeight = y.GetHeight()
		res.ShardID = y.GetShardID()
		res.BlkHash = y.Hash().String()
	} else {
		y := blockchain.NewBeaconBlock()
		y.UnmarshalJSON(msgPropose.Block)
		res.BlkHeight = y.GetHeight()
		res.ShardID = y.GetShardID()
		res.BlkHash = y.Hash().String()
	}
	return res
}

func ParseBFTVote(x *wire.MessageBFT) MsgVote {
	var msgVote blsbftv2.BFTVote
	err := json.Unmarshal(x.Content, &msgVote)
	if err != nil {
		fmt.Errorf("%v", err)
		return MsgVote{}
	}
	res := MsgVote{}
	res.MsgVote = &msgVote
	res.BlkHash = msgVote.BlockHash
	return res
}

func (f *OneBlockFork) CountingTimeSlot() {
	ticker := time.Tick(time.Second)
	oldTS := common.GetCurrentTimeSlot()
	curTS := int64(0)
	for range ticker {
		curTS = common.GetCurrentTimeSlot()
		if curTS > oldTS {
			fmt.Println("[fork] ---------------------------------------")
			fmt.Println("[fork] ------------ New time slot ------------")
			oldTS = curTS
		}
	}
}

func (f *OneBlockFork) WatchingChain(cID int) {
	x := map[uint64]uint64{}
	blkInfoByHash := map[string]BlkInfoByHash{} // <height_round_shardid>
	y := map[string]uint64{}                    // Number of vote <height_round>
	msgCh := f.BFTsCh[cID]
	for {
		select {
		case msg := <-msgCh:
			if bft, ok := msg.(*wire.MessageBFT); ok {
				if bft.Type == blsbftv2.MSG_PROPOSE {
					bftP := ParseBFTPropose(bft)
					if round, ok := x[bftP.BlkHeight]; ok {
						x[bftP.BlkHeight] = round + 1
					} else {
						x[bftP.BlkHeight] = 1
					}
					blkInfoByHash[bftP.BlkHash] = BlkInfoByHash{
						BlkHeight: bftP.BlkHeight,
						Round:     int(x[bftP.BlkHeight]),
						ShardID:   bftP.ShardID,
					}
				} else {
					bftV := ParseBFTVote(bft)
					blkInfo := blkInfoByHash[bftV.BlkHash]
					blkInfoKey := fmt.Sprintf("%v_%v", blkInfo.BlkHeight, blkInfo.Round)
					if total, ok := y[blkInfoKey]; ok {
						y[blkInfoKey] = total + 1
					} else {
						y[blkInfoKey] = 1
					}
				}
			}
		}
	}
}

type TriggerByBlock struct {
	S map[uint64]struct {
		TotalForkBlock uint64
		NextBlock      uint64
	}
}
