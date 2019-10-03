package main

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type HighwayConnector struct {
	host host.Host
	hmap *HighwayMap

	outPeers       chan peer.AddrInfo
	inPeers        chan peer.AddrInfo
	pendingInPeers []peer.AddrInfo
}

func NewHighwayConnector(host host.Host, hmap *HighwayMap) *HighwayConnector {
	hc := &HighwayConnector{
		host:     host,
		hmap:     hmap,
		outPeers: make(chan peer.AddrInfo, 1000),
	}

	// Register to receive notif when new connection is established
	host.Network().Notify((*notifiee)(hc))
	return hc
}

func (hc *HighwayConnector) Start() {
	for {
		select {
		case p := <-hc.outPeers:
			hc.host.Connect(context.Background(), p)

		case p := <-hc.inPeers:
			hc.checkInPeer(p)
		}
	}
}

func (hc *HighwayConnector) ConnectTo(p peer.AddrInfo) error {
	hc.outPeers <- p
	return nil
}

func (hc *HighwayConnector) checkInPeer(p peer.AddrInfo) {
	if hc.hmap.IsEnlisted(p) {
		// Update shards connected by this highway
		hc.hmap.ConnectToShardOfPeer(p)
	} else {
		// Add to pending if enlist message hasn't arrived
		hc.pendingInPeers = append(hc.pendingInPeers, p)
	}
}

type notifiee HighwayConnector

func (no *notifiee) Listen(network.Network, multiaddr.Multiaddr)      {}
func (no *notifiee) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (no *notifiee) Connected(n network.Network, c network.Conn) {
	go func() {
		no.inPeers <- peer.AddrInfo{
			ID:    c.RemotePeer(),
			Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
		}
	}()
}
func (no *notifiee) Disconnected(network.Network, network.Conn)   {}
func (no *notifiee) OpenedStream(network.Network, network.Stream) {}
func (no *notifiee) ClosedStream(network.Network, network.Stream) {}
