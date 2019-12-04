package rpcserver

import (
	"fmt"
)

type Handler struct {
	rpcServer *RpcServer
	// TODO using this param for support response all of HW peerID instead of default peerID for all shard
	// connector *route.Manager
}

func (s *Handler) GetPeers(
	req Request,
	res *Response,
) (
	err error,
) {
	fmt.Println(req)
	// Return default maps
	// if args[0] != "all" {
	// 	return nil, fmt.Errorf("Multi HW per shard is not supported in this time!")
	// }
	// s.connector.GetListHighways(pid peer.ID)
	res.PeerPerShard = map[string][]string{
		"all": []string{s.rpcServer.Config.IPFSAddr},
	}
	fmt.Println("Response", *res)
	return
}
