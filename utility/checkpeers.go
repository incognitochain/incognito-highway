package main

import (
	"fmt"
	"log"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func main() {
	filename := "keylist.json"
	keyListFromFile := loadKeylist(filename)

	r, err := getReport()
	if err != nil {
		log.Fatal(err)
	}

	// Beacon
	for i, val := range keyListFromFile.Bc {
		key := new(incognitokey.CommitteePublicKey)
		key.FromString(val.CommitteePubKey)
		miningKey := key.GetMiningKeyBase58("bls")

		if !found(miningKey, r.Chain.Peers[255]) {
			fmt.Printf("beacon, not connected to peer %d, mining key %s\n", i, miningKey)
		}
	}

	// Shards
	for sh, vals := range keyListFromFile.Sh {
		for i, val := range vals {
			key := new(incognitokey.CommitteePublicKey)
			key.FromString(val.CommitteePubKey)
			miningKey := key.GetMiningKeyBase58("bls")

			if !found(miningKey, r.Chain.Peers[sh]) {
				fmt.Printf("shard %d, not connected to peer %d, mining key %s\n", sh, i, miningKey)
			}
		}
	}
}
