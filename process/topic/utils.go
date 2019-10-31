package topic

import (
	"fmt"
	"highway/common"
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
		case CmdBlockBeacon, CmdBlkShardToBeacon, CmdCrossShard:
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
func GetCommitteeIDOfTopic(topic string) byte {
	topicElements := strings.Split(topic, "-")
	if len(topicElements) == 0 {
		return 0x00
	}
	bytesDecoded, _ := strconv.Atoi(topicElements[1])
	return byte(bytesDecoded)
}

func GetTopicForPubSub(msgType string, cID byte) string {
	if isBroadcastMessage(msgType) {
		return fmt.Sprintf("%s-%d-", msgType, cID)
	}
	return fmt.Sprintf("%s-%d-%s", msgType, cID, common.SelfID)
}

func GetTopicForPub(isHighway bool, msgType string, cID byte) string {
	commonTopic := GetTopicForPubSub(msgType, cID)
	if isHighway {
		return commonTopic + NODESUB
	} else {
		return commonTopic + NODEPUB
	}
}

func GetTopicForSub(isHighway bool, msgType string, cID byte) string {
	commonTopic := GetTopicForPubSub(msgType, cID)
	if !isHighway {
		return commonTopic + NODESUB
	} else {
		return commonTopic + NODEPUB
	}
}
