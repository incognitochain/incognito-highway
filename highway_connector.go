package main

import (
	"context"
	"encoding/json"
	logger "highway/customizelog"
	"highway/process"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

type HighwayConnector struct {
	host host.Host
	hmap *HighwayMap
	ps   *process.PubSubManager

	outPeers       chan peer.AddrInfo
	inPeers        chan peer.AddrInfo
	pendingInPeers []peer.AddrInfo
}

func NewHighwayConnector(host host.Host, hmap *HighwayMap, ps *process.PubSubManager) *HighwayConnector {
	hc := &HighwayConnector{
		host:     host,
		hmap:     hmap,
		ps:       ps,
		outPeers: make(chan peer.AddrInfo, 1000),
		inPeers:  make(chan peer.AddrInfo, 1000),
	}

	// Register to receive notif when new connection is established
	host.Network().Notify((*notifiee)(hc))

	// Start subscribing to receive enlist message from other highways
	hc.ps.NewMessage <- process.SubHandler{
		Topic:   "highway_enlist",
		Handler: hc.processEnlistMessage,
	}
	return hc
}

func (hc *HighwayConnector) Start() {
	for {
		select {
		case p := <-hc.outPeers:
			err := hc.dialAndEnlist(p)
			if err != nil {
				logger.Error(err, p)
			}

		case p := <-hc.inPeers:
			hc.checkInPeer(p)
		}
	}
}

func (hc *HighwayConnector) ConnectTo(p peer.AddrInfo) error {
	hc.outPeers <- p
	return nil
}

func (hc *HighwayConnector) processEnlistMessage(msg *pubsub.Message) {
	em := &enlistMessage{}
	err := json.Unmarshal(msg.Data, em)
	if err != nil {
		logger.Error(err)
		return
	}

	// Update supported shards of peer
	hc.hmap.AddPeer(em.Peer, em.SupportShards)
	hc.hmap.ConnectToShardOfPeer(em.Peer)
}

func (hc *HighwayConnector) dialAndEnlist(p peer.AddrInfo) error {
	err := hc.host.Connect(context.Background(), p)
	if err != nil {
		return errors.WithStack(err)
	}

	// Broadcast enlist message
	data := &enlistMessage{
		Peer: peer.AddrInfo{
			ID:    hc.host.ID(),
			Addrs: hc.host.Addrs(),
		},
		SupportShards: hc.hmap.Supports[hc.host.ID()],
	}
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "enlistMessage: %v", data)
	}
	if err := hc.ps.FloodMachine.Publish("highway_enlist", msg); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (hc *HighwayConnector) checkInPeer(p peer.AddrInfo) {
	// TODO(@0xakk0r0kamui): check highway's signature
	if hc.hmap.IsEnlisted(p) {
		logger.Info("Enlisted", p, hc.hmap.Supports[p.ID])
		// Update shards connected by this highway
		hc.hmap.ConnectToShardOfPeer(p)
	} else {
		logger.Info("Pending", p)
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

type enlistMessage struct {
	SupportShards []byte
	Peer          peer.AddrInfo
}
