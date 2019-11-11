package topic

import (
	"fmt"
	"highway/common"
	"testing"
)

func TestTopicManager_Init(t *testing.T) {
	tm := new(TopicManager)
	tm.Init()
	for msg, listPairSub := range tm.allTopicPairForNodeSub {
		listPairPub := tm.allTopicPairForNodePub[msg]
		flag := false
		for _, cID := range tm.allCommitteeID {
			pairSub := listPairSub[cID]
			pairPub := listPairPub[cID]
			if len(pairPub.Topic) > 0 || len(pairSub.Topic) > 0 {
				if !flag {
					fmt.Printf("Info of msg %v \n", msg)
					flag = true
				}
				fmt.Println("\tCommitteeID :", cID)
				if len(pairSub.Topic) > 0 {
					for i, t := range pairSub.Topic {
						fmt.Printf("\t\tTopic %v, Action %v\n", t, pairSub.Act[i])
					}
				}
				if len(pairPub.Topic) > 0 {
					for i, t := range pairPub.Topic {
						fmt.Printf("\t\tTopic %v, Action %v\n", t, pairPub.Act[i])
					}
				}
			}
		}
	}
}

func TestTopicManager_GetListTopicPairForNode(t *testing.T) {
	tm := new(TopicManager)
	tm.Init()
	res := tm.GetListTopicPairForNode(common.NORMAL, map[string][]int{
		CmdBFT:              []int{2},
		CmdBlockBeacon:      []int{2, 3, 255},
		CmdBlkShardToBeacon: []int{2, 3},
		CmdBlockShard:       []int{2, 3},
	})
	for _, pair := range res {
		fmt.Printf("%v %v\n", pair.Topic, pair.Act)
	}
}

func TestTopicManager_GetListSubTopicForHW(t *testing.T) {
	tm := new(TopicManager)
	tm.Init()
	tm.UpdateSupportShards([]byte{255})
	res := tm.GetListSubTopicForHW()
	fmt.Println("HW SUB----------------------")
	for _, topic := range res {
		fmt.Printf("\t%v\n", topic)
	}
}

func TestTopicManager_GetHWPubTopicsFromMsg(t *testing.T) {
	tm := new(TopicManager)
	tm.Init()
	for _, msg := range Message4Process {
		fmt.Printf("List Topic of Msg %v for HW Pub:\n", msg)
		for _, cID := range tm.allCommitteeID {
			tp := tm.GetHWPubTopicsFromMsg(msg, int(cID))
			if tp != "" {
				fmt.Printf("\t%v\n", tp)
			}
		}
	}
}

func TestTopicManager_GetHWPubTopicsFromHWSub(t *testing.T) {
	tm := new(TopicManager)
	tm.Init()
	tm.UpdateSupportShards([]byte{0})
	fmt.Printf("Highway supports cIDs: %v\n", tm.supportShards)
	listHWSub := tm.GetListSubTopicForHW()
	for _, subTopic := range listHWSub {
		pubFromSubTopic := tm.GetHWPubTopicsFromHWSub(subTopic)
		fmt.Printf("HW pub %v if receive %v\n", pubFromSubTopic, subTopic)
	}
}
