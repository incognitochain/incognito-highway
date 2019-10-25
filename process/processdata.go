package process

import (
	"fmt"
	logger "highway/customizelog"

	"github.com/pkg/errors"

	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	OneTopicOneData = iota
	OneTopicNData
	NTopicOneData
)

func PublishDataWithTopic(
	pubMachine *p2pPubSub.PubSub,
	pubTopics []string,
	listData [][]byte,
	mode int,
) error {
	if len(pubTopics) == 0 || len(listData) == 0 {
		return errors.New(fmt.Sprintf("Wrong input, list Topic (%v) or list Data (%v) is empty.", pubTopics, listData))
	}
	for i, data := range listData {
		switch mode {
		case OneTopicNData:
			err := pubMachine.Publish(pubTopics[0], data)
			if err != nil {
				return err
			}
		case OneTopicOneData:
			logger.Infof("Publish Topic %v mode %v", pubTopics[0], mode)
			if len(pubTopics) != len(listData) {
				logger.Errorf("Wrong input, in mode OneTopicOneData len of list Topic (%v) must equal len of list Data (%v).", pubTopics, listData)
				return errors.New(fmt.Sprintf("Wrong input, in mode OneTopicOneData len of list Topic (%v) must equal len of list Data (%v).", pubTopics, listData))
			}
			err := pubMachine.Publish(pubTopics[i], data)
			if err != nil {
				logger.Error(err)
				return err
			}
		case NTopicOneData:
			for _, pubTopic := range pubTopics {
				logger.Infof("Publish Topic %v mode %v", pubTopic, mode)
				err := pubMachine.Publish(pubTopic, data)
				if err != nil {
					return err
				}
			}
		default:
			return errors.New(fmt.Sprintf("Wrong mode, input mode %v", mode))
		}
	}
	return nil
}
