package main

import (
	"context"
	"encoding/json"
	logger "highway/customizelog"
	"highway/process"

	"github.com/incognitochain/incognito-chain/peerv2"
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

	outPeers chan peer.AddrInfo
}

func NewHighwayConnector(host host.Host, hmap *HighwayMap, ps *process.PubSubManager) *HighwayConnector {
	hc := &HighwayConnector{
		host:     host,
		hmap:     hmap,
		ps:       ps,
		outPeers: make(chan peer.AddrInfo, 1000),
	}

	// Register to receive notif when new connection is established
	host.Network().Notify((*notifiee)(hc))

	// Start subscribing to receive enlist message from other highways
	hc.ps.GRPCSpecSub <- process.SubHandler{
		Topic:   "highway_enlist",
		Handler: hc.enlistHighways,
	}

	// Subscribe to receive new committee
	// TODO(@0xbunyip): move logic updating committee to another object
	hc.ps.GRPCSpecSub <- process.SubHandler{
		Topic:   "chain_committee",
		Handler: hc.saveNewCommittee,
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
		}
	}
}

func (hc *HighwayConnector) ConnectTo(p peer.AddrInfo) error {
	hc.outPeers <- p
	return nil
}

func (hc *HighwayConnector) enlistHighways(sub *pubsub.Subscription) {
	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			logger.Error(err)
			continue
		}

		// TODO(@0xakk0r0kamui): check highway's signature in msg
		em := &enlistMessage{}
		if err := json.Unmarshal(msg.Data, em); err != nil {
			logger.Error(err)
			continue
		}

		// Update supported shards of peer
		hc.hmap.AddPeer(em.Peer, em.SupportShards)
		hc.hmap.ConnectToShardOfPeer(em.Peer)
	}
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

func (hc *HighwayConnector) saveNewCommittee(sub *pubsub.Subscription) {
	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			logger.Error(err)
			continue
		}

		comm := &peerv2.ChainCommittee{}
		if err := json.Unmarshal(msg.Data, comm); err != nil {
			logger.Error(err)
			continue
		}
	}
}

type notifiee HighwayConnector

func (no *notifiee) Listen(network.Network, multiaddr.Multiaddr)      {}
func (no *notifiee) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (no *notifiee) Connected(n network.Network, c network.Conn) {
	// TODO(@0xbunyip): check if highway or node connection
}
func (no *notifiee) Disconnected(network.Network, network.Conn)   {}
func (no *notifiee) OpenedStream(network.Network, network.Stream) {}
func (no *notifiee) ClosedStream(network.Network, network.Stream) {}

type enlistMessage struct {
	SupportShards []byte
	Peer          peer.AddrInfo
}
