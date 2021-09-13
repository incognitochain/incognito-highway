package route

import (
	"context"
	"encoding/json"
	"highway/common"
	"highway/process"
	"highway/proto"
	hmap "highway/route/hmap"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	pubsub "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// TODO(@0xbunyip): swap connector and client
// client provides rpc and connector manages connection
// also, move server out of connector
type Connector struct {
	host      host.Host
	hmap      *hmap.Map
	publisher Publisher

	hwc ConnKeeper
	hws *Server

	outPeers   chan peer.AddrInfo
	closePeers chan peer.ID
	stop       chan int

	masternode    peer.ID
	rpcUrl        string
	addrInfo      peer.AddrInfo
	supportShards []byte
	hwwhitelist   map[string]struct{}
}

func NewConnector(
	h host.Host,
	prtc *p2pgrpc.GRPCProtocol,
	hmap *hmap.Map,
	subHandlers chan process.SubHandler,
	publisher Publisher,
	masternode peer.ID,
	rpcUrl string,
	addrInfo peer.AddrInfo,
	supportShards []byte,
	whitelisthw map[string]struct{},
) *Connector {
	hc := &Connector{
		host:          h,
		hmap:          hmap,
		publisher:     publisher,
		hws:           NewServer(prtc, hmap), // GRPC server serving other highways
		hwc:           NewClient(prtc),       // GRPC clients to other highways
		outPeers:      make(chan peer.AddrInfo, 1000),
		closePeers:    make(chan peer.ID, 100),
		masternode:    masternode,
		rpcUrl:        rpcUrl,
		addrInfo:      addrInfo,
		supportShards: supportShards,
		stop:          make(chan int),
		hwwhitelist:   whitelisthw,
	}

	// Register to receive notif when new connection is established
	h.Network().Notify((*notifiee)(hc))

	// Start subscribing to receive enlist message from other highways
	subHandlers <- process.SubHandler{
		Topic:   "highway_enlist",
		Handler: hc.enlistHighways,
	}
	return hc
}

func (hc *Connector) GetHWClient(pid peer.ID) (proto.HighwayConnectorServiceClient, error) {
	return hc.hwc.GetClient(pid)
}

func (hc *Connector) Start() {
	enlistTimestep := time.NewTicker(common.BroadcastMsgEnlistTimestep)
	defer enlistTimestep.Stop()
	for {
		var err error
		select {
		case <-enlistTimestep.C:
			err = hc.enlist()

		case p := <-hc.outPeers:
			err = hc.dialAndEnlist(p)
			if err != nil {
				err = errors.WithMessagef(err, "peer: %+v", p)
			}

		case p := <-hc.closePeers:
			err = hc.closePeer(p)
			if err != nil {
				err = errors.WithMessagef(err, "peer: %+v", p)
			}

		case <-hc.stop:
			logger.Infof("Stopping connector loop")
			return
		}

		if err != nil {
			logger.Error(err)
		}
	}
}

func (hc *Connector) ConnectTo(p peer.AddrInfo) error {
	hc.outPeers <- p
	return nil
}

func (hc *Connector) Dial(p peer.AddrInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := hc.host.Connect(ctx, p)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (hc *Connector) CloseConnection(p peer.ID) {
	hc.closePeers <- p
}

func (hc *Connector) closePeer(p peer.ID) error {
	logger.Infof("Closing connection to peer %v", p)
	err := hc.host.Network().ClosePeer(p)
	err2 := hc.hwc.CloseConnection(p)
	if err != nil {
		return errors.WithStack(err)
	}
	if err2 != nil {
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
		em := &enlistMessage{}
		if msg.GetFrom().Pretty() == hc.host.ID().Pretty() {
			continue
		}
		if err := json.Unmarshal(msg.Data, em); err != nil {
			logger.Error(err)
			continue
		}

		if _, ok := hc.hwwhitelist[msg.GetFrom().String()]; !ok {
			logger.Infof("%v not from list", msg.GetFrom().String())
			hc.publisher.BlacklistPeer(msg.GetFrom())
			hc.hmap.RemoveRPCUrl(em.RPCUrl)
			continue
		}

		logger.Infof("Received highway_enlist msg: %+v", em)

		// Update supported shards of peer
		hc.hmap.AddPeer(em.Peer, em.SupportShards, em.RPCUrl)
	}
}

func (hc *Connector) enlist() error {
	// Broadcast enlist message
	data := &enlistMessage{
		Peer:          hc.addrInfo,
		SupportShards: hc.supportShards,
		RPCUrl:        hc.rpcUrl,
	}
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "enlistMessage: %v", data)
	}

	logger.Infof("Publishing msg highway_enlist: %s", msg)
	if err := hc.publisher.Publish("highway_enlist", msg); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (hc *Connector) dialAndEnlist(p peer.AddrInfo) error {
	if hc.host.Network().Connectedness(p.ID) != network.Connected {
		logger.Infof("Dialing to peer %+v", p)
		if err := hc.Dial(p); err != nil {
			return err
		}
		if hc.host.Network().Connectedness(p.ID) != network.Connected {
			if err := hc.enlist(); err != nil {
				return err
			}
		}
	}

	// Update list of connected shards
	hc.hmap.ConnectToShardOfPeer(p)

	// Publish msg enlist
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

type Publisher interface {
	Publish(topic string, msg []byte, opts ...pubsub.PubOpt) error
	BlacklistPeer(peer.ID)
}

type ConnKeeper interface {
	GetClient(peer.ID) (proto.HighwayConnectorServiceClient, error)
	CloseConnection(peerID peer.ID) error
	GetConnection(peerID peer.ID) (*grpc.ClientConn, error)
}
