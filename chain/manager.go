package chain

import (
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

type Manager struct {
}

func RegisterNotification(h host.Host) {
	m := &Manager{}
	h.Network().Notify((*notifiee)(m))
}

type notifiee Manager

func (no *notifiee) Listen(network.Network, multiaddr.Multiaddr)      {}
func (no *notifiee) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (no *notifiee) Connected(n network.Network, c network.Conn)      {}
func (no *notifiee) OpenedStream(network.Network, network.Stream)     {}
func (no *notifiee) ClosedStream(network.Network, network.Stream)     {}

func (no *notifiee) Disconnected(network.Network, network.Conn) {
}
