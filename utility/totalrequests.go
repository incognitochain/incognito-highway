package main

import (
	"fmt"
	"highway/common"
	"log"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func main() {
	filename := "key_s1"
	keyListFromFile := loadKeylist(filename)

	r, err := getReport()
	if err != nil {
		log.Fatal(err)
	}

	// Beacon
	sum := 0
	for _, v := range getTotalRequests(r, keyListFromFile.Bc) {
		sum += v
	}
	fmt.Println(filename)
	fmt.Printf("Total request beacon: %d\n", sum)

	// Shards
	for i, vals := range keyListFromFile.Sh {
		sum := 0
		for _, v := range getTotalRequests(r, vals) {
			sum += v
		}
		fmt.Printf("Total request shard %d: %d\n", i, sum)
	}
}

func getTotalRequests(r *Report, keys []common.Key) map[string]int {
	calls := map[string]int{}
	for _, key := range keys {
		k := new(incognitokey.CommitteePublicKey)
		k.FromString(key.CommitteePubKey)
		miningKey := k.GetMiningKeyBase58("bls")

		pid := getPeerIDFromMiningKey(r, miningKey)

		for _, reqs := range r.Chain.RequestPerPeer {
			for p, cnt := range reqs {
				if p == pid {
					calls[p] = calls[p] + cnt
				}
			}
		}
	}
	return calls
}

func getPeerIDFromMiningKey(r *Report, miningKey string) string {
	for _, peers := range r.Chain.Peers {
		for _, p := range peers {
			if p.Pubkey == miningKey {
				return p.ID
			}
		}
	}
	return ""
}
