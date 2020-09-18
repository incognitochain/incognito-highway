package datahandler

import (
	"highway/process/topic"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BlkBeaconHandler struct {
	FromNode bool
	PubSub   *libp2p.PubSub
}

func (handler *BlkBeaconHandler) HandleDataFromTopic(topicReceived string, dataReceived libp2p.Message) error {
	var topicPubs []string
	msgType := topic.GetMsgTypeOfTopic(topicReceived)
	cID := topic.GetCommitteeIDOfTopic(topicReceived)
	if handler.FromNode {
		topicPub := topic.Handler.GetHWPubSubOutSideFromMsg(msgType, cID)
		topicPubs = append(topicPubs, topicPub)
	} else {
		topicPubs = topic.Handler.GetHWPubTopicsFromMsg(msgType, cID)
	}
	logger.Debugf("[msgbeaconblk] Handle topic %v, isInside %v:", topicReceived, handler.FromNode)
	for _, topicPub := range topicPubs {
		logger.Debugf("[msgbeaconblk]\tPublish topic %v", topicPub)
		err := handler.PubSub.Publish(topicPub, dataReceived.GetData())
		if err != nil {
			logger.Errorf("Publish topic %v return error: %v", topicPub, err)
		}
	}
	return nil
}

func init() {
	cf := zap.NewDevelopmentConfig()
	cf.Level.SetLevel(zapcore.FatalLevel)
	l, _ := cf.Build()
	logger = l.Sugar()

	// chain.InitLogger(logger)
	// chaindata.InitLogger(logger)
	// process.InitLogger(logger)
	// topic.InitLogger(logger)
	// health.InitLogger(logger)
	// rpcserver.InitLogger(logger)
	// hmap.InitLogger(logger)
	InitLogger(logger)
}
