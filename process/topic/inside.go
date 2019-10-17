package topic

import (
	"errors"
	"fmt"
	"highway/common"
	logger "highway/customizelog"
)

// InsideTopic is topic inside a committee, using for communicate between proxy node and INC node
// MessageType is defined as type of message
// CommitteeID is defined as type of committee (0->254: shardID, 255: beacon)
// SelfID (optional) is ID of Proxy node
type InsideTopic struct {
	MessageType string
	CommitteeID byte
	SelfID      string
}

func (topic *InsideTopic) ToString() string {
	return fmt.Sprintf("%s-%x-%s", topic.MessageType, topic.CommitteeID, topic.SelfID)
}

func (topic *InsideTopic) FromMessageType(
	committeeID int,
	messageType string,
) error {
	if (committeeID) < 0 {
		logger.Infof("Candidate not found")
		return errors.New("Candidate not found")
	}
	topic.CommitteeID = byte(committeeID)
	// topic.CommitteeID = byte(0x00)
	if isBroadcastMessage(messageType) {
		topic.SelfID = ""
	} else {
		topic.SelfID = common.SelfID
	}
	//Validate correctness of messageType
	topic.MessageType = messageType
	return nil
}

func (topic *InsideTopic) GetTopic4ProxyPub() string {
	if topic.IsJustPubOrSub() {
		return topic.ToString() + "-nodesub"
	}
	return topic.ToString()
}

func (topic *InsideTopic) GetTopic4ProxySub() string {
	if IsJustPubOrSubMsg(topic.MessageType) {
		return topic.ToString() + "-nodepub"
	}
	return topic.ToString()
}

func GetTopic4ProxySub(
	committeeID byte,
	message string,
) string {
	topicGenerator := &InsideTopic{
		MessageType: message,
		CommitteeID: committeeID,
		SelfID:      common.SelfID,
	}
	if isBroadcastMessage(message) {
		topicGenerator.SelfID = ""
	}
	if IsJustPubOrSubMsg(topicGenerator.MessageType) {
		return topicGenerator.ToString() + "-nodepub"
	}
	return topicGenerator.ToString()
}

func GetTopic4ProxyPub(
	committeeID byte,
	message string,
) string {
	topicGenerator := &InsideTopic{
		MessageType: message,
		CommitteeID: committeeID,
		SelfID:      common.SelfID,
	}
	if isBroadcastMessage(message) {
		topicGenerator.SelfID = ""
	}
	if IsJustPubOrSubMsg(topicGenerator.MessageType) {
		return topicGenerator.ToString() + "-nodesub"
	}
	return topicGenerator.ToString()
}

func (topic *InsideTopic) IsJustPubOrSub() bool {
	return IsJustPubOrSubMsg(topic.MessageType)
}
