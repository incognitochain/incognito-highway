package process

import (
	"fmt"

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
			if len(pubTopics) != len(listData) {
				logger.Errorf("Wrong input, in mode OneTopicOneData len of list Topic (%v) must equal len of list Data (%v).", pubTopics, listData)
				return errors.New(fmt.Sprintf("Wrong input, in mode OneTopicOneData len of list Topic (%v) must equal len of list Data (%v).", pubTopics, listData))
			}
			logger.Infof("Publish Topic %v mode OneTopicOneData", pubTopics[i])
			err := pubMachine.Publish(pubTopics[i], data)
			if err != nil {
				logger.Errorf("Publish topic %v return err %v", pubTopics[i], err)
				return err
			}
		case NTopicOneData:
			logger.Infof("Publish Topic %v mode NTopicOneData", pubTopics)
			for _, pubTopic := range pubTopics {
				err := pubMachine.Publish(pubTopic, data)
				if err != nil {
					logger.Errorf("Publish topic %v return err %v", pubTopic, err)
					return err
				}
			}
		default:
			return errors.New(fmt.Sprintf("Wrong mode, input mode %v", mode))
		}
	}
	return nil
}
