package simulateutils

import (
	"encoding/json"
	"fmt"
	"highway/common"
	"highway/process/topic"
	"sync"
	"time"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/incognitochain/incognito-chain/blockchain"
	chaincommon "github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/blsbftv2"
	"github.com/incognitochain/incognito-chain/wire"
)

type BasicForkScene struct {
	TotalForkBlock uint64
	NextBlock      uint64
}

type TriggerByBlock struct {
	S              map[uint64]BasicForkScene
	triggerBlkHash map[string]struct{}
	sync.RWMutex
}

func NewTriggerByBlock() TriggerByBlock {
	return TriggerByBlock{
		S:              map[uint64]BasicForkScene{},
		triggerBlkHash: map[string]struct{}{},
	}
}

type SimpleTrigger struct {
	ByBlock        map[uint64]BasicForkScene
	ByTxID         map[string]BasicForkScene
	triggerBlkHash map[string]struct{}
	sync.RWMutex
}

func NewSimpleTrigger() *SimpleTrigger {
	return &SimpleTrigger{
		ByBlock:        map[uint64]BasicForkScene{},
		ByTxID:         map[string]BasicForkScene{},
		triggerBlkHash: map[string]struct{}{},
	}
}

func (s *TriggerByBlock) IsTrigger(msg wire.Message) (bool, BasicForkScene) {
	if msgBFT, ok := msg.(*wire.MessageBFT); ok {
		if msgBFT.Type == blsbftv2.MSG_PROPOSE {
			bftP := ParseBFTPropose(msgBFT)
			if (bftP.BlkHeight > 3420) && (bftP.BlkHeight < 3430) {
				s.Lock()
				s.triggerBlkHash[bftP.BlkHash] = struct{}{}
				s.Unlock()
				return true, BasicForkScene{
					TotalForkBlock: 2,
					NextBlock:      2,
				}
			}
			if fs, ok := s.S[bftP.BlkHeight]; ok {
				s.Lock()
				s.triggerBlkHash[bftP.BlkHash] = struct{}{}
				s.Unlock()
				return ok, fs
			}
		} else {
			bftV := ParseBFTVote(msgBFT)
			s.RLock()
			if _, ok := s.triggerBlkHash[bftV.BlkHash]; ok {
				s.RUnlock()
				return true, BasicForkScene{
					TotalForkBlock: 2,
					NextBlock:      2,
				}
			}
			s.RUnlock()
		}
	}
	return false, BasicForkScene{}
}

func (s *SimpleTrigger) IsTrigger(msg wire.Message) (bool, BasicForkScene) {
	if msgBFT, ok := msg.(*wire.MessageBFT); ok {
		if msgBFT.Type == blsbftv2.MSG_PROPOSE {
			bftP := ParseBFTPropose(msgBFT)
			// if (bftP.BlkHeight > 3420) && (bftP.BlkHeight < 3430) {
			// 	s.Lock()
			// 	s.triggerBlkHash[bftP.BlkHash] = struct{}{}
			// 	s.Unlock()
			// 	return true, BasicForkScene{
			// 		TotalForkBlock: 2,
			// 		NextBlock:      2,
			// 	}
			// }
			//Check blockheight
			s.Lock()
			if fs, ok := s.ByBlock[bftP.BlkHeight]; ok {
				s.triggerBlkHash[bftP.BlkHash] = struct{}{}
				s.Unlock()
				return ok, fs
			}
			for k, v := range s.ByTxID {
				if CheckIfExistTxIDInBlk(msgBFT, k) {
					s.triggerBlkHash[bftP.BlkHash] = struct{}{}
					s.Unlock()
					return true, v
				}
			}
			s.Unlock()
		} else {
			bftV := ParseBFTVote(msgBFT)
			s.RLock()
			if _, ok := s.triggerBlkHash[bftV.BlkHash]; ok {
				s.RUnlock()
				return true, BasicForkScene{
					TotalForkBlock: 2,
					NextBlock:      2,
				}
			}
			s.RUnlock()
		}
	}
	return false, BasicForkScene{}
}

type OneBlockFork struct {
	CTable        CommitteeTable
	Scenes        ScenarioBasic
	BS            TriggerByBlock
	PBFTsCh       map[int]chan MsgBFT
	VBFTsCh       map[int]chan MsgBFT
	KeysCh        map[int]chan interface{}
	SupportShard  []byte
	StartSignal   []chan struct{}
	StopSignal    []chan struct{}
	CommitteeInfo *CommitteeTable
	PubSub        *libp2p.PubSub
	status        string
}

func (f *OneBlockFork) PBFTInbox(cID int) chan MsgBFT {
	return f.PBFTsCh[cID]
}

func (f *OneBlockFork) VBFTInbox(cID int) chan MsgBFT {
	return f.VBFTsCh[cID]
}

func (f *OneBlockFork) KeyInbox(cID int) chan interface{} {
	return f.KeysCh[cID]
}

func (f *OneBlockFork) SetScenes(sType SceneType, sByte []byte, cID int) error {
	err := f.Scenes.SetScenes(sType, sByte, cID)
	return err
}

func (f *OneBlockFork) GetScenes(SceneType) interface{} {
	return nil
}

func (f *OneBlockFork) GetPubGroupWithMsg(wire.Message) map[string][]string {
	return nil
}

func (f *OneBlockFork) IsTrigger(msg wire.Message, cID int) bool {
	f.Scenes.Lock.RLock()
	simpleScene, ok := f.Scenes.Scenes[cID]
	if !ok {
		return ok
	}
	f.Scenes.Lock.RUnlock()
	forked, _ := simpleScene.IsTrigger(msg)
	return forked
}

// func

func (f *OneBlockFork) Start() {
	f.status = "running"
	f.StartSignal = make([]chan struct{}, len(f.SupportShard))
	f.StopSignal = make([]chan struct{}, len(f.SupportShard))
	for i, cID := range f.SupportShard {
		f.StartSignal[i] = make(chan struct{})
		f.StopSignal[i] = make(chan struct{})
		// go f.WatchingChain(int(cID), f.StartSignal[i], f.StopSignal[i])
		go f.MakeItFork(int(cID))
	}
	// go f.CountingTimeSlot()

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
	ticker := time.Tick(200 * time.Millisecond)
	oldTS := common.GetCurrentTimeSlot()
	curTS := int64(0)
	// processingBlock := uint64(0)
	// nextBlock := uint64(1)
	for range ticker {
		curTS = common.GetCurrentTimeSlot()
		if curTS > oldTS {
			fmt.Println("[fork] ---------------------------------------")
			fmt.Println("[fork] ------------ New time slot ------------")
			oldTS = curTS
			for _, tmp := range f.StartSignal {
				tmp <- struct{}{}
			}
		}
	}
}

// CheckIfItFork This function just received BFT Propose and check if it trigger any fork scene
// Not handle vote msg
// TODO rename
// Flow:
// CheckIfItFork send list msg propose to PublishMachine and return
// PublishMachine make fork done and call CheckIfItFork
func (f *OneBlockFork) CheckIfItFork(
	cID int,
	wantedHeight uint64,
	pPubInfoCh chan PPublishInfo,
) error {
	msgProposeMap := map[int]MsgBFT{} //key <Round>
	pbftCh := f.PBFTInbox(cID)
	blkInfoMap := map[int]BlkInfoByHeight{}
	blkInfoByHash := map[string]BlkInfoByHash{} // <height_round>
	round := 1
	fork := false
	fs := BasicForkScene{}
	phase := "LISTENING"
	_ = phase
	// startTS := common.GetCurrentTimeSlot()
	pInfo := PPublishInfo{
		StartTimeSlot: 0,
		MsgBFTByRound: map[int]MsgBFT{},
		Height:        wantedHeight,
		ForkScene:     nil,
	}
	for msgPBFT := range pbftCh {
		msg := msgPBFT.Msg
		bft := msg.(*wire.MessageBFT)
		bftP := ParseBFTPropose(bft)
		if bftP.BlkHeight < wantedHeight {
			continue
		}
		fmt.Printf("[debugfork] Received msg proposes: cID %v Height %v Hash %v \n", bftP.ShardID, bftP.BlkHeight, bftP.BlkHash)
		if round == 1 {
			pInfo.StartTimeSlot = common.GetCurrentTimeSlot()
			// fork, fs = f.BS.IsTrigger(msg)
			f.Scenes.Lock.RLock()
			simpleScene, ok := f.Scenes.Scenes[cID]
			f.Scenes.Lock.RUnlock()
			if !ok {
				err := fmt.Errorf("Not support this cID %v", cID)
				panic(err)
			}
			if wantedHeight != 0 {
				fmt.Printf("[debugfork] %v %v \n", wantedHeight, bftP.BlkHeight)
				simpleScene.RLock()
				if _, ok := simpleScene.ByBlock[wantedHeight]; !ok {
					fmt.Println("[debugfork] Outdated heights, update wanted height")
					wantedHeight = 0
					pInfo.Height = wantedHeight
				}
				simpleScene.RUnlock()
			}
			fork, fs = simpleScene.IsTrigger(msg)
			if !fork {
				pInfo.ForkScene = nil
			} else {
				pInfo.ForkScene = &fs
			}
		}
		pInfo.MsgBFTByRound[round] = msgPBFT

		if fork {
			if (wantedHeight != bftP.BlkHeight) && (wantedHeight != 0) {
				fmt.Printf("Broken, received height %v, wanted height %v\n", bftP.BlkHeight, wantedHeight)
				return fmt.Errorf("Broken, received height %v, wanted height %v", bftP.BlkHeight, wantedHeight)
			}
		}
		pInfo.Height = bftP.BlkHeight
		msgProposeMap[round] = msgPBFT
		blkInfoMap[round] = BlkInfoByHeight{
			BlkHash: bftP.BlkHash,
			Round:   round,
			ShardID: bftP.ShardID,
			Propose: msgPBFT,
		}
		blkInfoByHash[bftP.BlkHash] = BlkInfoByHash{
			BlkHeight: bftP.BlkHeight,
			Round:     round,
			ShardID:   bftP.ShardID,
		}
		fmt.Printf("Chose block height %v cID %v blkHash %v", bftP.BlkHeight, bftP.ShardID, bftP.BlkHash)
		round++
		if fork {
			if uint64(len(msgProposeMap)) == fs.TotalForkBlock {
				// SendItToPublishMachine()
				fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
				fmt.Println(pInfo)
				fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
				pPubInfoCh <- pInfo
				return nil
			}
			continue
		}
		pPubInfoCh <- pInfo
		// } else {
		// 	// SendItToPublishMachine()
		// 	fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		// 	fmt.Println(pInfo)
		// 	fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		// 	pPubInfoCh <- pInfo
		return nil
		// }

	}
	return nil
}

func (f *OneBlockFork) GetNextPropose(cID int, stTS, stIdx, curTS int64) (string, int) {
	f.CommitteeInfo.lock.RLock()
	idx := int(stIdx+curTS-stTS) % f.CommitteeInfo.CommitteeSize[byte(cID)]
	f.CommitteeInfo.lock.RUnlock()
	pk := f.CommitteeInfo.PubKeyBySID[byte(cID)][idx]
	return pk, idx
}

func (f *OneBlockFork) PublishToPK(topicBFT, pk string, msg wire.Message) string {
	data, _ := common.EncodeMessage(msg)
	newTopic := topic.SwitchTopicPrivate(topicBFT, pk)
	f.PubSub.Publish(newTopic, []byte(data))
	fmt.Printf("[debugbft] Publish to topic %v v\n ", newTopic)
	return "ei"
}

func (f *OneBlockFork) PublishToCID(cID int, topicBFT string, msg wire.Message) string {
	data, _ := common.EncodeMessage(msg)
	for _, pk := range f.CommitteeInfo.PubKeyBySID[byte(cID)] {
		newTopic := topic.SwitchTopicPrivate(topicBFT, pk)
		err := f.PubSub.Publish(newTopic, []byte(data))
		fmt.Printf("[debugbft] Publish to topic %v error %v publisher %v\n ", newTopic, err, &f.PubSub)
	}
	return "ei"
}

func (f *OneBlockFork) PublishToCIDButNot(cID int, topicBFT string, pkIdxs []int, msg wire.Message) string {
	data, _ := common.EncodeMessage(msg)
	for idx, pk := range f.CommitteeInfo.PubKeyBySID[byte(cID)] {
		tmp := false
		for _, ignoreID := range pkIdxs {
			if idx == ignoreID {
				tmp = true
				break
			}
		}
		if tmp {
			continue
		}
		newTopic := topic.SwitchTopicPrivate(topicBFT, pk)
		f.PubSub.Publish(newTopic, []byte(data))
		fmt.Printf("[debugbft] Publish to topic %v\n ", newTopic)
	}
	return "ei"
}

// MakeItFork This function receive batch of finalize BFT msgs of a blk height
// Not process/parse bft propose, work with vote and redirect message for make fork
// TODO rename
func (f *OneBlockFork) MakeItFork(cID int) {
	oldTS := common.GetCurrentTimeSlot()
	curTS := int64(0)
	mapBFTToPub := map[int]MsgBFT{}
	blkInfoMap := map[int]*BlkInfoByHeight{}
	blkInfoByHash := map[string]int{}
	curHeight := uint64(0)
	curfs := &BasicForkScene{}
	forkStatus := "WaitPropose"
	pInfo := PPublishInfo{
		StartTimeSlot: 0,
		MsgBFTByRound: map[int]MsgBFT{},
		Height:        0,
		ForkScene:     nil,
	}
	pPubInfoCh := make(chan PPublishInfo, 1000)
	go f.CheckIfItFork(
		cID,
		0,
		pPubInfoCh,
	)
	round := 1
	for {
		curTS = common.GetCurrentTimeSlot()
		if curTS > oldTS {
			fmt.Println("[fork] ---------------------------------------")
			fmt.Println("[fork] ------------ New time slot ------------")
			fmt.Println("CID:", cID, "  ", f.Scenes.Scenes[cID])
			fmt.Printf("[fork] CID %v current TS: %v, current Height: %v, ForkStatus: %v\n", cID, curTS, curHeight, forkStatus)
			oldTS = curTS
			switch forkStatus {
			case "WaitPropose":
				time.Sleep((time.Duration(common.TIMESLOT)/2 + 1) * time.Second)
				select {
				case pInfo = <-pPubInfoCh:
					mapBFTToPub = pInfo.MsgBFTByRound
					curHeight = pInfo.Height
					if pInfo.ForkScene != nil {
						round = int(pInfo.ForkScene.TotalForkBlock)
						curfs = pInfo.ForkScene
					} else {
						round = 1
					}
					forkStatus = "PublishPropose"
				default:
					continue
				}
			case "PublishPropose":
				if len(mapBFTToPub) > 0 {
					fmt.Printf("TS %v Publish bft msg to cid %v topic %v round %v msg %v \n", curTS, cID, mapBFTToPub[round].Topic, round, mapBFTToPub[round].Msg)
					for k, v := range mapBFTToPub {
						fmt.Printf("K: %v Value %v\n", k, v)
					}
					f.PublishToCID(cID, mapBFTToPub[round].Topic, mapBFTToPub[round].Msg)
					bftP := ParseBFTPropose(mapBFTToPub[round].Msg.(*wire.MessageBFT))
					blkInfoMap[round] = &BlkInfoByHeight{
						Propose: mapBFTToPub[round],
						BlkHash: bftP.BlkHash,
						Vote:    []wire.Message{},
						Round:   round,
						ShardID: bftP.ShardID,
					}
					blkInfoByHash[bftP.BlkHash] = round
					curHeight = bftP.BlkHeight
					// if err == nil {
					// 	delete(mapBFTToPub, i)
					// }
					round--
					if round == 0 {
						forkStatus = "WaitVote"
					}
				}
			case "WaitVote":
				fmt.Printf("TS %v\n", curTS)
				ticker := time.NewTicker(10 * time.Millisecond)
				out := false
				for {
					select {
					case msgBFT := <-f.VBFTsCh[cID]:
						msg := msgBFT.Msg
						bft := msg.(*wire.MessageBFT)
						fmt.Printf("[debugfork] CID %v Received msg vote %v\n", cID, chaincommon.HashB(bft.Content))
						bftV := ParseBFTVote(bft)
						fmt.Printf("[debugfork] Parse result %v\n", bftV.BlkHash)
						voteR, ok := blkInfoByHash[bftV.BlkHash]
						if (ok) && (blkInfoMap[voteR].BlkHash == bftV.BlkHash) {
							blkInfoMap[voteR].Vote = append(blkInfoMap[voteR].Vote, msg)
							if (voteR == 1) && (len(blkInfoMap[voteR].Vote) > len(f.CommitteeInfo.PubKeyBySID[byte(cID)])/3*2) {
								fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
								forkStatus = "PublishVote"
								out = true
								break
							}
						}
					case <-ticker.C:
						if common.GetCurrentTimeSlot() > curTS {
							fmt.Println("Force waiting for end of timeslot", common.GetCurrentTimeSlot(), curTS)
							out = true
							break
						}
					}
					if out {
						break
					}
				}
			case "PublishVote":
				if pInfo.ForkScene != nil {
					pk, idx := f.GetNextPropose(
						cID,
						pInfo.StartTimeSlot,
						int64(blkInfoMap[1].Propose.ProposeIndex),
						curTS,
					)
					fmt.Printf("[fork] Next propose for fork scence (%v, %v): %v %v", pInfo.ForkScene.NextBlock, pInfo.ForkScene.TotalForkBlock, pk, idx)
					for _, msgV := range blkInfoMap[int(curfs.NextBlock)].Vote {
						f.PublishToPK(
							blkInfoMap[int(curfs.NextBlock)].Propose.Topic,
							pk,
							msgV,
						)
					}
					for i := pInfo.ForkScene.TotalForkBlock; i > 0; i-- {
						listVoteInfo := blkInfoMap[int(i)]
						topicBFT := listVoteInfo.Propose.Topic
						for _, vote := range listVoteInfo.Vote {
							f.PublishToCIDButNot(
								cID,
								topicBFT,
								[]int{idx},
								vote,
							)
						}
						delete(blkInfoMap, int(i))
					}
				} else {
					for _, msgV := range blkInfoMap[1].Vote {
						f.PublishToCID(
							cID,
							blkInfoMap[1].Propose.Topic,
							msgV,
						)
					}
					delete(blkInfoMap, 1)
				}
				go f.CheckIfItFork(
					cID,
					curHeight+1,
					pPubInfoCh,
				)
				forkStatus = "WaitPropose"
				continue
			}

			// for _, tmp := range f.StartSignal {
			// 	tmp <- struct{}{}
			// }
		}
	}
}

func (f *OneBlockFork) WatchingChain(cID int, start chan struct{}, stop chan struct{}) {
	// msgProposeMap := map[string]wire.Message{} //key <cID>_<blockHeight>_<Round>
	// msgVoteMap := map[string][]wire.Message{}  //key <cID>_<blockHeight>_<Round>
	blkInfoByHeight := map[uint64][]BlkInfoByHeight{}
	blkInfoByHash := map[string]BlkInfoByHash{} // <height_round_shardid>
	y := map[string]uint64{}                    // Number of vote <height_round>
	msgCh := f.PBFTsCh[cID]
	max := uint64(1)
	// newBlock := uint64(0)
	newRound := uint64(0)
	// ticker := time.NewTicker(100 * time.Millisecond)
	// oldTS := common.GetCurrentTimeSlot()
	// curTS := int64(0)
	for {
		select {
		case msgStruct := <-msgCh:
			msg := msgStruct.Msg
			senderPubkey := msg.(*wire.MessageBFT).PeerID
			if bft, ok := msg.(*wire.MessageBFT); ok {
				if bft.Type == blsbftv2.MSG_PROPOSE {
					fmt.Printf("[debugfork] CID %v Received msg propose %v\n", cID, chaincommon.HashB(bft.Content))
					bftP := ParseBFTPropose(bft)
					fmt.Printf("[debugfork] Parse result %v\n", bftP.BlkHash)
					if bftP.BlkHeight > max {
						// newBlock = bftP.BlkHeight
						newRound = 1
						max = bftP.BlkHeight
					} else {
						// newBlock = 0
						newRound++
					}
					// msgPMKey := fmt.Sprintf("%v_%v_%v", cID, max, newRound)
					// msgProposeMap[msgPMKey] = msg
					if infos, ok := blkInfoByHeight[bftP.BlkHeight]; !ok {
						blkInfoByHeight[bftP.BlkHeight] = []BlkInfoByHeight{
							{
								BlkHash: bftP.BlkHash,
								Round:   1,
								ShardID: bftP.ShardID,
								Propose: msgStruct,
							},
						}
					} else {
						blkInfoByHeight[bftP.BlkHeight] = append(
							infos,
							BlkInfoByHeight{
								BlkHash: bftP.BlkHash,
								Round:   len(infos) + 1,
								ShardID: bftP.ShardID,
								Propose: msgStruct,
							},
						)
					}
					blkInfoByHash[bftP.BlkHash] = BlkInfoByHash{
						BlkHeight: bftP.BlkHeight,
						Round:     len(blkInfoByHeight[bftP.BlkHeight]),
						ShardID:   bftP.ShardID,
					}
				} else {
					fmt.Printf("[debugfork] CID %v Received msg vote %v\n", cID, chaincommon.HashB(bft.Content))
					bftV := ParseBFTVote(bft)
					fmt.Printf("[debugfork] Parse result %v\n", bftV.BlkHash)
					blkInfo := blkInfoByHash[bftV.BlkHash]
					blkInfoKey := fmt.Sprintf("%v_%v", blkInfo.BlkHeight, blkInfo.Round)
					if total, ok := y[blkInfoKey]; ok {
						y[blkInfoKey] = total + 1
					} else {
						y[blkInfoKey] = 1
					}
					blkInfoByHeight[blkInfo.BlkHeight][blkInfo.Round-1].Vote = append(blkInfoByHeight[blkInfo.BlkHeight][blkInfo.Round-1].Vote, msg)
					fmt.Printf("[debugfork] cID %v Added vote %v %v\n", cID, blkInfoKey, y[blkInfoKey])
				}
				data, _ := common.EncodeMessage(msg)
				for _, dstPKey := range f.CommitteeInfo.GetKeysByKey(senderPubkey) {
					newTopic := topic.SwitchTopicPrivate(msgStruct.Topic, dstPKey)
					f.PubSub.Publish(newTopic, []byte(data))
					fmt.Printf("[debugbft] BFT of cID %v Publish to topic %v v\n ", cID, newTopic)
				}
			}
		case <-start:
			fmt.Printf("[fork] CID %v Received:\n", cID)
			fmt.Printf("[fork] BlkHeight: %v \n", max-1)
			if infos, ok := blkInfoByHeight[max-1]; ok {
				for _, info := range infos {
					fmt.Printf("[fork] BlkHash: %v; ", info.BlkHash)
					fmt.Printf("#vote %v\n", len(info.Vote))
				}
			}
			fmt.Printf("[fork] BlkHeight: %v \n", max)
			if infos, ok := blkInfoByHeight[max]; ok {
				for _, info := range infos {
					fmt.Printf("[fork] BlkHash: %v; ", info.BlkHash)
					fmt.Printf("#vote %v\n", len(info.Vote))
				}
			}
			<-start
		}
	}
}
