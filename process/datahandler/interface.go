package datahandler

import (
	libp2p "github.com/libp2p/go-libp2p-pubsub"
)

type Handler interface {
	HandleNewSubs(subs *libp2p.Subscription)
}

type DataHandler interface {
	HandleDataFromTopic(topic string, dataReceived libp2p.Message) error
}
