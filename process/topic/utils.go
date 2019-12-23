package topic

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var TypeOfTopicProcessor map[string]byte

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

func InitTypeOfProcessor() {
	TypeOfTopicProcessor = map[string]byte{}
	for _, mess := range Message4Process {
		switch mess {
		case CmdPeerState:
			TypeOfTopicProcessor[mess] = ProcessAndPublishAfter
		case CmdBlockBeacon, CmdBlkShardToBeacon, CmdCrossShard, CmdBlockShard, CmdTx, CmdCustomToken, CmdPrivacyCustomToken:
			TypeOfTopicProcessor[mess] = ProcessAndPublish
		default:
			TypeOfTopicProcessor[mess] = DoNothing
		}
	}
}

func IsJustPubOrSubMsg(msg string) bool {
	switch msg {
	case CmdPeerState, CmdBlkShardToBeacon, CmdBlockBeacon, CmdCrossShard, CmdBlockShard:
		return true
	default:
		return false
	}
}

func GetTypeOfProcess(topic string) byte {
	topicElements := strings.Split(topic, "-")
	if len(topicElements) == 0 {
		return WTFisThis
	}
	return TypeOfTopicProcessor[topicElements[0]]
}

// GetMsgTypeOfTopic handle error later
func GetMsgTypeOfTopic(topic string) string {
	topicElements := strings.Split(topic, "-")
	if len(topicElements) == 0 {
		return ""
	}
	return topicElements[0]
}

// GetCommitteeIDOfTopic handle error later TODO handle error pls
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

// FromInsideToOutsideTopic
// input a topic inside committee, such as crossshard-2-<self ID>-<optinal>
// output a topic for communication between HW, like crossshard-2
func FromInsideToOutsideTopic(topic string) string {
	msgType := GetMsgTypeOfTopic(topic)
	cID := GetCommitteeIDOfTopic(topic)
	return fmt.Sprintf("%v-%v", msgType, cID)
}
