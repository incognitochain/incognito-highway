package main

import (
	"fmt"
	"highway/rpcserver"
)

func main() {
	rpcConf := &rpcserver.RpcServerConfig{
		Port: 9330,
	}
	server, err := rpcserver.NewRPCServer(rpcConf, nil)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	server.Start()
	select {}
}
