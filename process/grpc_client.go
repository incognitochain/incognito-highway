package process

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
)

type GRPCService_Client struct {
	P2Pgrpc *p2pgrpc.GRPCProtocol
}

func (self *GRPCService_Client) ProxyRegister(ctx context.Context, peerID peer.ID, pubkey string) ([]string, error) {
	return nil, nil
	// grpcConn, err := self.p2pgrpc.Dial(ctx, peerID, grpc.WithInsecure(), grpc.WithBlock())
	// if err != nil {
	// 	return nil, err
	// }
	// client := NewProxyRegisterServiceClient(grpcConn)
	// reply, err := client.ProxyRegister(ctx, &ProxyRegisterMsg{CommitteePublicKey: "11234567890"})
	// if err != nil {
	// 	log.Fatalln(err)
	// 	return nil, err
	// }
	// return reply.Topics, nil
}
