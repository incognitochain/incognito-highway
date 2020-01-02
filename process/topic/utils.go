package topic

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func isBroadcastMessage(message string) bool {
	if message == CmdBFT {
		return true
	}
	return false
}

func isValidMessage(message string) bool {
	sort.Slice(Message4Process, func(i, j int) bool {
		return Message4Process[i] < Message4Process[j]
	})
	idx := sort.SearchStrings(Message4Process, message)
	if (idx < 0) || (idx >= len(Message4Process)) {
		return false
	}
	return true
}

func IsJustPubOrSubMsg(msg string) bool {
	switch msg {
	case CmdPeerState, CmdBlkShardToBeacon, CmdBlockBeacon, CmdCrossShard, CmdBlockShard:
		return true
	default:
		return false
	}
}

// GetMsgTypeOfTopic handle error later
func GetMsgTypeOfTopic(topic string) string {
	topicElements := strings.Split(topic, "-")
	if len(topicElements) == 0 {
		return ""
	}
	return topicElements[0]
}

func GetCommitteeIDOfTopic(topic string) int {
	topicElements := strings.Split(topic, "-")
	if len(topicElements) == 0 {
		return -1
	}
	if topicElements[1] == "" {
		return NoCIDInTopic
	}
	cID, _ := strconv.Atoi(topicElements[1])
	return cID
}

func getTopicForPubSub(msgType string, cID int, selfID string) string {
	if isBroadcastMessage(msgType) {
		return fmt.Sprintf("%s-%d-", msgType, cID)
	}
	if cID == NoCIDInTopic {
		return fmt.Sprintf("%s--%s", msgType, selfID)
	}
	return fmt.Sprintf("%s-%d-%s", msgType, cID, selfID)
}

func getTopicForPub(side, msgType string, cID int, selfID string) string {
	commonTopic := getTopicForPubSub(msgType, cID, selfID)
	if side == HIGHWAYSIDE {
		return commonTopic + NODESUB
	} else {
		return commonTopic + NODEPUB
	}
}

func getTopicForSub(side, msgType string, cID int, selfID string) string {
	commonTopic := getTopicForPubSub(msgType, cID, selfID)
	if side == NODESIDE {
		return commonTopic + NODESUB
	} else {
		return commonTopic + NODEPUB
	}
}
