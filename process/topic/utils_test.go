package topic

import (
	"fmt"
	"testing"
)

func TestGetCommitteeIDOfTopic(t *testing.T) {
	tm := TopicManager{}
	tm.Init()
	for msg, listPair := range tm.getAllTopicPairForNode(true) {
		fmt.Printf("Msg %v:\n", msg)
		for CID, pair := range listPair {
			if len(pair.Topic) == 0 {
				continue
			}
			for _, tp := range pair.Topic {
				cID := GetCommitteeIDOfTopic(tp)
				fmt.Printf("Topics: %v, cID: %v\n", tp, cID)
				switch msg {
				case CmdBlockBeacon, CmdBlkShardToBeacon:
					if cID != noCIDInTopic {
						t.Errorf("GetCommitteeIDOfTopic(%v) = %v, want %v", tp, cID, noCIDInTopic)
					}
				case CmdCrossShard:
					//Don't know how to check
					continue
				default:
					if cID != int(CID) {
						t.Errorf("GetCommitteeIDOfTopic(%v) = %v, want %v", tp, cID, CID)
					}
				}
			}
		}
	}
}
