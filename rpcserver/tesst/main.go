package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

func B2ImN(seed []byte) *big.Int {
	x := big.NewInt(0)
	x.SetBytes(common.HashB(seed))
	for x.Cmp(crypto.ECDSACurve.Params().N) != -1 {
		x.SetBytes(common.HashB(x.Bytes()))
	}
	return x
}

func newPriKey(seed []byte) *ecdsa.PrivateKey {
	priKey := new(ecdsa.PrivateKey)
	priKey.Curve = crypto.ECDSACurve
	priKey.D = B2ImN(seed)
	priKey.PublicKey.X, priKey.PublicKey.Y = priKey.Curve.ScalarBaseMult(priKey.D.Bytes())
	return priKey //, priKey.PublicKey
}

func testKey() {
	seed := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	for i := 0; i < 10; i++ {
		ecdsaPrivate := newPriKey(append(seed, byte(i)))
		pri, pub, err := crypto.ECDSAKeyPairFromKey(ecdsaPrivate)
		if err != nil {
			fmt.Println(err)
		}
		_ = pub
		ctx := context.Background()
		opts := []libp2p.Option{
			libp2p.Identity(pri),
		}
		priBytes, err := pri.Bytes()
		if err != nil {
			fmt.Println(err)
		}
		priString := crypto.ConfigEncodeKey(priBytes)

		p2pHost, err := libp2p.New(ctx, opts...)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%v %v\n", priString, p2pHost.ID())
	}
}

func main() {
	testKey()
	// rpcConf := &rpcserver.RpcServerConfig{
	// 	Port: 9330,
	// }
	// server, err := rpcserver.NewRPCServer(rpcConf, nil)
	// if err != nil {
	// 	fmt.Printf("%v\n", err)
	// }
	// server.Start()
	// select {}
}
