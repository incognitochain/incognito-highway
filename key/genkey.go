package key

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

func bytesToIntmN(seed []byte) *big.Int {
	x := big.NewInt(0)
	x.SetBytes(common.HashB(seed))
	for x.Cmp(crypto.ECDSACurve.Params().N) != -1 {
		x.SetBytes(common.HashB(x.Bytes()))
	}
	return x
}

func priKeyFromSeed(seed []byte) *ecdsa.PrivateKey {
	priKey := new(ecdsa.PrivateKey)
	priKey.Curve = crypto.ECDSACurve
	priKey.D = bytesToIntmN(seed)
	priKey.PublicKey.X, priKey.PublicKey.Y = priKey.Curve.ScalarBaseMult(priKey.D.Bytes())
	return priKey //, priKey.PublicKey
}

func peerInfofromSeed(seed []byte) (peer.ID, crypto.PrivKey, error) {
	ecdsaPriKey := priKeyFromSeed(seed)
	pri, pub, err := crypto.ECDSAKeyPairFromKey(ecdsaPriKey)
	if err != nil {
		return peer.ID(""), nil, err
	}
	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return peer.ID(""), nil, err
	}
	return pid, pri, nil
}

func GenWhiteList(
	seed []byte,
	selfIdx int,
	lenlist int,
) (
	map[string]struct{},
	crypto.PrivKey,
	error,
) {
	whileList := map[string]struct{}{}
	var selfPri crypto.PrivKey
	for i := 0; i < lenlist; i++ {
		newSeed := append(seed, byte(i))
		pID, pri, err := peerInfofromSeed(newSeed)
		if err != nil {
			return nil, nil, err
		}
		whileList[pID.String()] = struct{}{}
		if i == selfIdx {
			selfPri = pri
		}
	}
	return whileList, selfPri, nil
}
