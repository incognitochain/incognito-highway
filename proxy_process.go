package main

import (
	"bufio"
	"github.com/libp2p/go-libp2p-core/network"
)

func (s *ProxyHost) ProcessConnection() {
	s.Host.SetStreamHandler(s.GetProxyStreamProtocolID(), func(stream network.Stream) {
		_ = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	})
}
