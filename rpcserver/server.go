package rpcserver

import (
	"fmt"
	"net"
	"net/rpc"
)

type RpcServer struct {
	// peers    map[string]*peer // list peers which are still pinging to bootnode continuously
	// peersMtx sync.Mutex
	server *rpc.Server
	Config *RpcServerConfig // config for RPC server
}

func NewRPCServer(
	conf *RpcServerConfig,
) (
	*RpcServer,
	error,
) {
	rpcServer := new(RpcServer)
	rpcServer.server = rpc.NewServer()
	rpcServer.Config = conf
	return rpcServer, nil
}

func (rpcServer *RpcServer) Start() error {
	handler := &Handler{rpcServer}
	rpcServer.server.Register(handler)
	listenAddr := fmt.Sprintf(":%d", rpcServer.Config.Port)
	listenner, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("listen in address %v error: %v\n", listenAddr, err)
		return err
	}
	rpcServer.server.Accept(listenner)
	listenner.Close()
	return nil
}
