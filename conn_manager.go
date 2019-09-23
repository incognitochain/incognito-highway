package main

import (
	"bufio"
	"context"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"highway/p2p"
)

type PeerConn struct {
	RemotePeer *p2p.Peer
	RW         *bufio.ReadWriter
}

type ConnManager struct {
	Host     host.Host
	peerTag  map[peer.ID]connmgr.TagInfo
	peerConn map[peer.ID][]PeerConn
	actionCh chan func()
}

func (s *ConnManager) Start() {
	s.actionCh = make(chan func())
	for {
		select {
		case action := <-s.actionCh:
			action()
		}
	}
}

func (s *ConnManager) AddPeerConnection(id peer.ID, conn PeerConn) {
	s.actionCh <- func() {
		s.peerConn[id] = append(s.peerConn[id], conn)
	}
}

func (s *ConnManager) TagPeer(peer.ID, string, int)                          {}
func (s *ConnManager) UntagPeer(p peer.ID, tag string)                       {}
func (s *ConnManager) UpsertTag(p peer.ID, tag string, upsert func(int) int) {}
func (s *ConnManager) GetTagInfo(p peer.ID) *connmgr.TagInfo {
	return nil
}
func (s *ConnManager) TrimOpenConns(ctx context.Context) {}
func (s *ConnManager) Notifee() network.Notifiee         { return nil }
func (s *ConnManager) Protect(id peer.ID, tag string)    {}
func (s *ConnManager) Unprotect(id peer.ID, tag string) (protected bool) {
	return true
}
func (s *ConnManager) Close() error {
	return nil
}
