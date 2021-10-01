package topic

import (
	"fmt"
	"highway/common"
	"highway/proto"
	"sync"
)

type listPairByCID map[byte]proto.MessageTopicPair

// TODO @0xakk0r0kamui remove this global param in next pull request
var Handler TopicManager

type TopicManager struct {
	allTopicPairForNodeSub map[string]listPairByCID
	rwLockTopicNodeSub     *sync.RWMutex
	allTopicPairForNodePub map[string]listPairByCID
	rwLockTopicNodePub     *sync.RWMutex
	isInit                 bool
	allCommitteeID         []byte
	supportShards          []byte
	selfID                 string
}

func (topicManager *TopicManager) UpdateSupportShards(supportShards []byte) {
	for _, s := range supportShards {
		if common.HasValuesAt(topicManager.supportShards, s) != -1 {
			continue
		}
		topicManager.supportShards = append(topicManager.supportShards, s)
	}
}

func (topicManager *TopicManager) Init(selfID string) {
	topicManager.selfID = selfID
	if topicManager.isInit {
		return
	}
	topicManager.allCommitteeID = make([]byte, common.NumberOfShard+1)
	for i := 0; i < common.NumberOfShard; i++ {
		topicManager.allCommitteeID[i] = byte(i)
	}
	topicManager.allCommitteeID[common.NumberOfShard] = common.BEACONID
	topicManager.rwLockTopicNodeSub = new(sync.RWMutex)
	topicManager.rwLockTopicNodePub = new(sync.RWMutex)
	topicManager.allTopicPairForNodeSub = topicManager.getAllTopicPairForNode(false)
	topicManager.allTopicPairForNodePub = topicManager.getAllTopicPairForNode(true)
	topicManager.isInit = true
}

func (topicManager *TopicManager) getAllTopicPairForNode(
	forPub bool,
) map[string]listPairByCID {
	res := map[string]listPairByCID{}
	for _, msg := range Message4Process {
		listPair := map[byte]proto.MessageTopicPair{}
		for _, s := range topicManager.allCommitteeID {
			listPair[s] = topicManager.getTopicPairForNode(msg, forPub, s)
		}
		res[msg] = listPair
	}
	return res
}

func (topicManager *TopicManager) GetAllTopicOutsideForHW() []string {
	allTopics := []string{}
	for _, msg := range Message4Process {
		listTopics := []string{}
		for _, cID := range topicManager.supportShards {
			pair := getTopicPairOutsideForHW(msg, cID)
			listTopics = append(listTopics, pair.Topic...)
		}

		for _, topicOutside := range listTopics {
			if (len(topicOutside) > 0) && (common.HasStringAt(allTopics, topicOutside) == -1) {
				allTopics = append(allTopics, topicOutside)
			}
		}
	}
	return allTopics
}

func getTopicPairOutsideForHW(
	msgType string,
	cID byte,
) proto.MessageTopicPair {
	listTopic := []string{}
	listAction := []proto.MessageTopicPair_Action{}
	act := proto.MessageTopicPair_PUBSUB
	cIDint := int(cID)
	isHasTopics := true
	switch msgType {
	case CmdBFT, CmdPeerState:
	case CmdBlockBeacon:
		cIDint = NoCIDInTopic
	case CmdBlkShardToBeacon, CmdFinishSync:
		if cID == common.BEACONID {
			cIDint = NoCIDInTopic
		} else {
			isHasTopics = false
		}
	case CmdBlockShard, CmdTx, CmdCustomToken, CmdPrivacyCustomToken, CmdCrossShard:
		if cID == common.BEACONID {
			isHasTopics = false
		}
	default:
		isHasTopics = false
	}
	if isHasTopics {
		listTopic = append(listTopic, getTopicOutSideFromMsg(msgType, cIDint))
		listAction = append(listAction, act)
	}
	pair := proto.MessageTopicPair{
		Message: msgType,
		Topic:   listTopic,
		Act:     listAction,
	}
	return pair
}

func (topicManager *TopicManager) getTopicPairForNode(
	msgType string,
	forPub bool,
	cID byte,
) proto.MessageTopicPair {
	listTopic := []string{}
	listAction := []proto.MessageTopicPair_Action{}
	switch msgType {
	case CmdBFT:
		listTopic = append(listTopic, getTopicForPubSub(msgType, int(cID), topicManager.selfID))
		listAction = append(listAction, proto.MessageTopicPair_PUBSUB)
	case CmdPeerState:
		if forPub {
			listTopic = append(listTopic, getTopicForPub(NODESIDE, msgType, int(cID), topicManager.selfID))
			listAction = append(listAction, proto.MessageTopicPair_PUB)
		} else {
			listTopic = append(listTopic, getTopicForSub(NODESIDE, msgType, int(cID), topicManager.selfID))
			listAction = append(listAction, proto.MessageTopicPair_SUB)
		}
	case CmdBlockShard, CmdTx, CmdCustomToken, CmdPrivacyCustomToken:
		if cID != common.BEACONID {
			if forPub {
				listTopic = append(listTopic, getTopicForPub(NODESIDE, msgType, int(cID), topicManager.selfID))
				listAction = append(listAction, proto.MessageTopicPair_PUB)
			} else {
				listTopic = append(listTopic, getTopicForSub(NODESIDE, msgType, int(cID), topicManager.selfID))
				listAction = append(listAction, proto.MessageTopicPair_SUB)
			}
		} else {
			if !forPub {
				listTopic = append(listTopic, getTopicForSub(NODESIDE, msgType, NoCIDInTopic, topicManager.selfID))
				listAction = append(listAction, proto.MessageTopicPair_SUB)
			}
		}
	case CmdBlockBeacon:
		if forPub {
			if cID == common.BEACONID {
				listTopic = append(listTopic, getTopicForPub(NODESIDE, msgType, NoCIDInTopic, topicManager.selfID))
				listAction = append(listAction, proto.MessageTopicPair_PUB)
			}
		} else {
			listTopic = append(listTopic, getTopicForSub(NODESIDE, msgType, NoCIDInTopic, topicManager.selfID))
			listAction = append(listAction, proto.MessageTopicPair_SUB)
		}
	case CmdBlkShardToBeacon, CmdFinishSync:
		if forPub {
			if cID != common.BEACONID {
				listTopic = append(listTopic, getTopicForPub(NODESIDE, msgType, NoCIDInTopic, topicManager.selfID))
				listAction = append(listAction, proto.MessageTopicPair_PUB)
			}
		} else {
			if cID == common.BEACONID {
				listTopic = append(listTopic, getTopicForSub(NODESIDE, msgType, NoCIDInTopic, topicManager.selfID))
				listAction = append(listAction, proto.MessageTopicPair_SUB)
			}
		}
	case CmdCrossShard:
		if cID != common.BEACONID {
			if forPub {
				for _, s := range topicManager.allCommitteeID {
					if (s != byte(cID)) && (s != common.BEACONID) {
						listTopic = append(listTopic, getTopicForPub(NODESIDE, msgType, int(s), topicManager.selfID))
						listAction = append(listAction, proto.MessageTopicPair_PUB)
					}
				}
			} else {
				listTopic = append(listTopic, getTopicForSub(NODESIDE, msgType, int(cID), topicManager.selfID))
				listAction = append(listAction, proto.MessageTopicPair_SUB)
			}
		}
	}
	pair := proto.MessageTopicPair{
		Message: msgType,
		Topic:   listTopic,
		Act:     listAction,
	}
	return pair
}

func (topicManager *TopicManager) GetListTopicPairForMonitor() []*proto.MessageTopicPair {
	res := []*proto.MessageTopicPair{}
	// listTopicsPubsub := topicManager.GetAllTopicOutsideForHW()
	// for
	for _, msg := range Message4Process {
		listTopics := []string{}
		listAcc := []proto.MessageTopicPair_Action{}
		for _, cID := range topicManager.supportShards {
			pair := getTopicPairOutsideForHW(msg, cID)
			for _, tp := range pair.Topic {
				if (len(tp) > 0) && (common.HasStringAt(listTopics, tp) == -1) {
					listTopics = append(listTopics, tp)
					listAcc = append(listAcc, proto.MessageTopicPair_SUB)
				}
			}
		}
		if len(listTopics) > 0 {
			res = append(res, &proto.MessageTopicPair{
				Message: msg,
				Topic:   listTopics,
				Act:     listAcc,
			})
		}
	}
	return res
}

func (topicManager *TopicManager) GetListTopicPairForNode(
	level byte,
	msgAndCID map[string][]int,
) []*proto.MessageTopicPair {
	res := []*proto.MessageTopicPair{}
	for msg, listCID := range msgAndCID {
		topics := []string{}
		actions := []proto.MessageTopicPair_Action{}
		if level <= lvlAllowPubOfMsg[msg] {
			for _, cID := range listCID {
				if common.HasValuesAt(topicManager.supportShards, byte(cID)) == -1 {
					continue
				}
				topicManager.rwLockTopicNodePub.RLock()
				for i, topicSub := range topicManager.allTopicPairForNodePub[msg][byte(cID)].Topic {
					if common.HasStringAt(topics, topicSub) != -1 {
						continue
					}
					topics = append(topics, topicSub)
					actions = append(actions, topicManager.allTopicPairForNodePub[msg][byte(cID)].Act[i])
				}
				topicManager.rwLockTopicNodePub.RUnlock()
			}
		}
		if level <= lvlAllowSubOfMsg[msg] {
			for _, cID := range listCID {
				if common.HasValuesAt(topicManager.allCommitteeID, byte(cID)) == -1 {
					continue
				}
				topicManager.rwLockTopicNodeSub.RLock()
				for i, topicSub := range topicManager.allTopicPairForNodeSub[msg][byte(cID)].Topic {
					if common.HasStringAt(topics, topicSub) != -1 {
						continue
					}
					topics = append(topics, topicSub)
					actions = append(actions, topicManager.allTopicPairForNodeSub[msg][byte(cID)].Act[i])
				}
				topicManager.rwLockTopicNodeSub.RUnlock()
			}
		}

		if len(topics) == 0 {
			continue
		}
		pair := &proto.MessageTopicPair{
			Message: msg,
			Topic:   topics,
			Act:     actions,
		}
		res = append(res, pair)
	}
	return res
}

func (topicManager *TopicManager) GetListSubTopicForHW() []string {
	res := []string{}
	locker := topicManager.rwLockTopicNodePub
	allTopic := topicManager.allTopicPairForNodePub
	locker.RLock()
	for _, listPair := range allTopic {
		for _, cID := range topicManager.supportShards {
			for _, t := range listPair[cID].Topic {
				if common.HasStringAt(res, t) == -1 {
					res = append(res, t)
				}
			}
		}
	}
	locker.RUnlock()
	return res
}

func (topicManager *TopicManager) GetHWPubTopicsFromMsg(msg string, cID int) []string {
	if cID == NoCIDInTopic {
		for _, cid := range topicManager.supportShards {
			if pair, ok := topicManager.allTopicPairForNodeSub[msg][cid]; ok {
				if len(pair.Topic) > 0 {
					return pair.Topic
				}
			}
		}
	} else {
		if pair, ok := topicManager.allTopicPairForNodeSub[msg][byte(cID)]; ok {
			if len(pair.Topic) > 0 {
				return pair.Topic
			}
		}
	}
	return []string{}
}

func getTopicOutSideFromMsg(msg string, cID int) string {
	if cID == NoCIDInTopic {
		return fmt.Sprintf("%v--", msg)
	}
	return fmt.Sprintf("%v-%d-", msg, cID)
}

func (topicManager *TopicManager) GetHWPubSubOutSideFromMsg(msg string, cID int) string {
	return getTopicOutSideFromMsg(msg, cID)
}
