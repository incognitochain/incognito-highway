package main

import (
	"context"
	p2pPubSub "github.com/libp2p/go-libp2p-pubsub"
	"highway/p2p"
)

func ProcessConnection(s *p2p.Host) {
	pubsub, err := p2pPubSub.NewGossipSub(context.Background(), s.Host)
	must(err)

	//subscribe to beacon
	pubsub.Subscribe("beacon")

}
