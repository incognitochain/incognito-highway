package route

import (
	"context"
	"encoding/json"
	"highway/process"
	"highway/proto"
	hmap "highway/route/hmap"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

// TODO(@0xbunyip): swap connector and client
// client provides rpc and connector manages connection
// also, move server out of connector
type Connector struct {
	host host.Host
	hmap *hmap.Map
	ps   *process.PubSubManager
	hwc  *Client
	hws  *Server

	outPeers chan peer.AddrInfo

	masternode peer.ID
	rpcUrl     string
}

func NewConnector(
	h host.Host,
	prtc *p2pgrpc.GRPCProtocol,
	hmap *hmap.Map,
	ps *process.PubSubManager,
	masternode peer.ID,
	rpcUrl string,
) *Connector {
	hc := &Connector{
		host:       h,
		hmap:       hmap,
		ps:         ps,
		hws:        NewServer(prtc, hmap), // GRPC server serving other highways
		hwc:        NewClient(prtc),       // GRPC clients to other highways
		outPeers:   make(chan peer.AddrInfo, 1000),
		masternode: masternode,
		rpcUrl:     rpcUrl,
	}

	// Register to receive notif when new connection is established
	h.Network().Notify((*notifiee)(hc))

	// Start subscribing to receive enlist message from other highways
	hc.ps.SubHandlers <- process.SubHandler{
		Topic:   "highway_enlist",
		Handler: hc.enlistHighways,
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
		logger.Infof("Received highway_enlist msg: %+v", em)

		// Update supported shards of peer
		hc.hmap.AddPeer(em.Peer, em.SupportShards, em.RPCUrl)
	}
}

func (hc *Connector) dialAndEnlist(p peer.AddrInfo) error {
	if hc.host.Network().Connectedness(p.ID) == network.Connected {
		return nil
	}

	logger.Infof("Dialing to peer %+v", p)
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
		RPCUrl:        hc.rpcUrl,
	}
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "enlistMessage: %v", data)
	}
	logger.Infof("Publishing msg highway_enlist: %s", msg)
	if err := hc.ps.FloodMachine.Publish("highway_enlist", msg); err != nil {
		return errors.WithStack(err)
	}

	// Update list of connected shards
	hc.hmap.ConnectToShardOfPeer(p)
	return nil
}

type notifiee Connector

func (no *notifiee) Listen(network.Network, multiaddr.Multiaddr)      {}
func (no *notifiee) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (no *notifiee) Connected(n network.Network, c network.Conn) {
	// TODO(@0xbunyip): check if highway or node connection
	// log.Println("route/manager: new conn")
}
func (no *notifiee) Disconnected(network.Network, network.Conn)   {}
func (no *notifiee) OpenedStream(network.Network, network.Stream) {}
func (no *notifiee) ClosedStream(network.Network, network.Stream) {}

type enlistMessage struct {
	SupportShards []byte
	Peer          peer.AddrInfo
	RPCUrl        string
}
