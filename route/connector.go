package route

import (
	"context"
	"encoding/json"
	"highway/process"
	"highway/proto"
	"log"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

type Connector struct {
	host host.Host
	hmap *Map
	ps   *process.PubSubManager
	hwc  *Client
	hws  *Server

	outPeers chan peer.AddrInfo

	masternode peer.ID
}

func NewConnector(
	h host.Host,
	prtc *p2pgrpc.GRPCProtocol,
	hmap *Map,
	ps *process.PubSubManager,
	masternode peer.ID,
) *Connector {
	hc := &Connector{
		host:       h,
		hmap:       hmap,
		ps:         ps,
		hws:        NewServer(prtc), // GRPC server serving other highways
		hwc:        NewClient(prtc), // GRPC clients to other highways
		outPeers:   make(chan peer.AddrInfo, 1000),
		masternode: masternode,
	}

	// Register to receive notif when new connection is established
	h.Network().Notify((*notifiee)(hc))

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

func (hc *Connector) GetHWClient(pid peer.ID) (proto.HighwayConnectorServiceClient, error) {
	return hc.hwc.GetClient(pid)
}

func (hc *Connector) Start() {
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

func (hc *Connector) ConnectTo(p peer.AddrInfo) error {
	hc.outPeers <- p
	return nil
}

func (hc *Connector) Dial(p peer.AddrInfo) error {
	err := hc.host.Connect(context.Background(), p)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (hc *Connector) enlistHighways(sub *pubsub.Subscription) {
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

func (hc *Connector) dialAndEnlist(p peer.AddrInfo) error {
	err := hc.Dial(p)
	if err != nil {
		return err
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

func (hc *Connector) saveNewCommittee(sub *pubsub.Subscription) {
	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		logger.Info("Received new committee")
		if err != nil {
			logger.Error(err)
			continue
		}

		// TODO(@0xbunyip): check if msg.From can be manipulated by forwarder
		if peer.ID(msg.From) != hc.masternode {
			from := peer.IDB58Encode(peer.ID(msg.From))
			exp := peer.IDB58Encode(hc.masternode)
			logger.Warnf("Received NewCommittee from unauthorized source, expect from %+v, got from %+v, data %+v", from, exp, msg.Data)
			continue
		}

		logger.Info("Saving new committee")
		comm := &incognitokey.ChainCommittee{}
		if err := json.Unmarshal(msg.Data, comm); err != nil {
			logger.Error(err)
			continue
		}

		// TOOD(@0xbunyip): update chain committee to ChainData here
	}
}

type notifiee Connector

func (no *notifiee) Listen(network.Network, multiaddr.Multiaddr)      {}
func (no *notifiee) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (no *notifiee) Connected(n network.Network, c network.Conn) {
	// TODO(@0xbunyip): check if highway or node connection
	log.Println("route/manager: new conn")
}
func (no *notifiee) Disconnected(network.Network, network.Conn)   {}
func (no *notifiee) OpenedStream(network.Network, network.Stream) {}
func (no *notifiee) ClosedStream(network.Network, network.Stream) {}

type enlistMessage struct {
	SupportShards []byte
	Peer          peer.AddrInfo
}
