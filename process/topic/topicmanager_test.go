package topic

import (
	"fmt"
	"highway/common"
	"testing"
)

func TestTopicManager_Init(t *testing.T) {
	tm := new(TopicManager)
	tm.Init("aa")
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
	tm.Init("aa")
	tm.UpdateSupportShards(tm.allCommitteeID)
	res := tm.GetListTopicPairForNode(common.COMMITTEE, map[string][]int{
		CmdBFT:                []int{2},
		CmdBlockBeacon:        []int{2, 3, 255},
		CmdBlkShardToBeacon:   []int{2, 3},
		CmdBlockShard:         []int{2, 3},
		CmdTx:                 []int{2, 3},
		CmdPrivacyCustomToken: []int{2, 3},
		CmdCustomToken:        []int{2, 3},
	})
	for _, pair := range res {
		fmt.Printf("%v %v\n", pair.Topic, pair.Act)
	}
}

func TestTopicManager_GetListSubTopicForHW(t *testing.T) {
	tm := new(TopicManager)
	tm.Init("aa")
	tm.UpdateSupportShards([]byte{0, 1, 2, 3, 4, 5, 6, 7, 255})
	res := tm.GetListSubTopicForHW()
	fmt.Println("HW SUB----------------------")
	for _, topic := range res {
		fmt.Printf("\t%v\n", topic)
	}
}

func TestTopicManager_GetHWPubTopicsFromMsg(t *testing.T) {
	tm := new(TopicManager)
	tm.Init("aa")
	for _, msg := range Message4Process {
		fmt.Printf("List Topic of Msg %v for HW Pub:\n", msg)
		for _, cID := range tm.allCommitteeID {
			tp := tm.GetHWPubTopicsFromMsg(msg, int(cID))
			if len(tp) != 0 {
				fmt.Printf("\t%v\n", tp)
			}
		}
	}
}

func Test_getTopicPairOutsideForHW(t *testing.T) {
	supportShards := []byte{0, 1, 2, 3, 4, 5, 6, 7, 255}
	for _, msgType := range Message4Process {
		fmt.Println("Message: ", msgType)
		for _, cID := range supportShards {
			fmt.Printf("\tCommittee ID: %v\n", cID)
			topicPair := getTopicPairOutsideForHW(msgType, cID)
			fmt.Printf("\tHW pubsub topic %v\n", topicPair.Topic)
		}
	}
}

func TestTopicManager_GetAllTopicOutsideForHW(t *testing.T) {
	tm := new(TopicManager)
	tm.Init("aa")
	tm.UpdateSupportShards([]byte{255, 0, 1, 2, 3, 4, 5, 6, 7})
	fmt.Println(tm.GetAllTopicOutsideForHW())
}

func TestTopicManager_GetListTopicPairForMonitor(t *testing.T) {
	tm := new(TopicManager)
	tm.Init("aa")
	tm.UpdateSupportShards([]byte{255, 0, 1, 2, 3, 4, 5, 6, 7})
	got := tm.GetListTopicPairForMonitor()
	for _, v := range got {
		fmt.Println(v)
	}
}
