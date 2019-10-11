package process

import (
	"context"
	logger "highway/customizelog"
	"highway/process/topic"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
)

var GlobalPubsub PubSubManager

<<<<<<< HEAD
type SubHandler struct {
	Topic   string
	Handler func(*p2pPubSub.Message)
}

type PubSubManager struct {
	FloodMachine   *p2pPubSub.PubSub
	GossipMachine  *p2pPubSub.PubSub
	NewMessage     chan SubHandler
	insideMessage  []string
	outsideMessage []string
	ForwardNow     chan p2pPubSub.Message
	Msgs           []*p2pPubSub.Subscription
=======
// type Config struct {
// 	Suppo
// }

type PubSubManager struct {
	SupportShards        []byte
	FloodMachine         *p2pPubSub.PubSub
	GossipMachine        *p2pPubSub.PubSub
	RegisterMessage      chan string
	OutSideMessage       chan string
	followedTopic        []string
	outsideMessage       []string
	ForwardNow           chan p2pPubSub.Message
	Msgs                 []*p2pPubSub.Subscription
	SpecialPublishTicker *time.Ticker
>>>>>>> version-0.1-pubsub
}

func InitPubSub(s host.Host, supportShards []byte) error {
	ctx := context.Background()
	var err error
	GlobalPubsub.FloodMachine, err = p2pPubSub.NewFloodSub(ctx, s)
	if err != nil {
		return err
	}
	// GlobalPubsub.GossipMachine, err = p2pPubSub.NewGossipSub(ctx, s)
	// if err != nil {
	// 	return err
	// }
<<<<<<< HEAD
	GlobalPubsub.NewMessage = make(chan SubHandler, 100)
=======
	GlobalPubsub.RegisterMessage = make(chan string)
>>>>>>> version-0.1-pubsub
	GlobalPubsub.ForwardNow = make(chan p2pPubSub.Message)
	GlobalPubsub.Msgs = make([]*p2pPubSub.Subscription, 0)
	GlobalPubsub.SpecialPublishTicker = time.NewTicker(5 * time.Second)
	GlobalPubsub.SupportShards = supportShards
	topic.InitTypeOfProcessor()
	initGlobalParams()
	// done := make(chan bool)
	// go func() {
	// for {
	// select {
	// case <-done:
	// return
	// case t := <-ticker.C:
	// fmt.Println("Tick at", t)
	// }
	// }
	// }()
	return nil
}

func (pubsub *PubSubManager) WatchingChain() {
	for {
		select {
<<<<<<< HEAD
		case subHandler := <-pubsub.NewMessage:
			subch, err := pubsub.FloodMachine.Subscribe(subHandler.Topic)
=======
		case newTopic := <-pubsub.RegisterMessage:
			subch, err := pubsub.FloodMachine.Subscribe(newTopic)
			pubsub.followedTopic = append(pubsub.followedTopic, newTopic)
>>>>>>> version-0.1-pubsub
			if err != nil {
				logger.Info(err)
				continue
			}
<<<<<<< HEAD
			logger.Infof("Success subscribe topic %v\n", subHandler.Topic)
			pubsub.Msgs = append(pubsub.Msgs, subch)
			go pubsub.handleNewMsg(subch, subHandler.Handler)
=======
			typeOfProcessor := topic.GetTypeOfProcess(newTopic)
			logger.Infof("Topic %v, Type of processor %v", newTopic, typeOfProcessor)

			logger.Infof("Success subscribe topic %v, Type of process %v", newTopic, typeOfProcessor)
			pubsub.Msgs = append(pubsub.Msgs, subch)
			go pubsub.handleNewMsg(subch, typeOfProcessor)
		case <-pubsub.SpecialPublishTicker.C:
			for committeeID, committeeState := range AllPeerState {
				for _, stateData := range committeeState {
					PeriodicalPublish(pubsub.FloodMachine, topic.CmdPeerState, committeeID, stateData)
				}
			}
>>>>>>> version-0.1-pubsub
		}

	}
}

<<<<<<< HEAD
func (pubsub *PubSubManager) handleNewMsg(sub *p2pPubSub.Subscription, handler func(*p2pPubSub.Message)) {
	for {
		data, err := sub.Next(context.Background())
		// fmt.Println("~~~~~~~~~~", err, "~~~~~~~~~~", data, "~~~~~~~~~~")
		//TODO implement GossipSub with special topic
		if (err == nil) && (data != nil) {
			// err = pubsub.FloodMachine.Publish(sub.Topic(), data.GetData())
			if err == nil {
				// logger.Infof("Topic: %v, data: %v\n", sub.Topic(), string(data.Data))
			} else {
				logger.Infof("Publish topic %v failed, err: %v\n", sub.Topic(), err)
=======
func (pubsub *PubSubManager) handleNewMsg(sub *p2pPubSub.Subscription, typeOfProcessor byte) {
	for {
		data, err := sub.Next(context.Background())
		//TODO implement GossipSub with special topic
		if (err == nil) && (data != nil) {
			switch typeOfProcessor {
			case topic.DoNothing:
				logger.Infof("Receive data from topic %v DoNothing", sub.Topic())
				continue
			case topic.ProcessAndPublishAfter:
				// logger.Infof("Receive data ProcessAndPublishAfter") //, data.GetData())
				// logger.Info(CommitteePubkeyByPeerID)
				go UpdatePeerState(CommitteePubkeyByPeerID[data.GetFrom()], data.GetData())
			case topic.ProcessAndPublish:
				go ProcessNPublishDataFromTopic(pubsub.FloodMachine, sub.Topic(), data.GetData(), pubsub.SupportShards)
			default:
				return
>>>>>>> version-0.1-pubsub
			}
			handler(data)
		}
	}
}

func (pubsub *PubSubManager) hasTopic(receivedTopic string) bool {
	for _, flTopic := range pubsub.followedTopic {
		if receivedTopic == flTopic {
			return true
		}
	}
	return false
}
