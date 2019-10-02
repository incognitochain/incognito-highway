package main

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	host "github.com/libp2p/go-libp2p-host"
)

type HighwayConnector struct {
	host host.Host

	outPeers chan peer.AddrInfo
}

func NewHighwayConnector(host host.Host) *HighwayConnector {
	return &HighwayConnector{
		host:     host,
		outPeers: make(chan peer.AddrInfo, 1000),
	}
}

func (hc *HighwayConnector) Start() {
	for {
		select {
		case p := <-hc.outPeers:
			hc.host.Connect(context.Background(), p)
		}
	}
}

func (hc *HighwayConnector) ConnectTo(p peer.AddrInfo) error {
	hc.outPeers <- p
	return nil
}
