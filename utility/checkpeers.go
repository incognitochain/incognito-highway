package main

import (
	"encoding/json"
	"fmt"
	"highway/common"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type Report struct {
	Chain struct {
		Peers map[int][]struct {
			Pubkey string
		} `json:"peers"`
	} `json:"chain"`
}

func get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return body, nil
}

func getReport() (*Report, error) {
	url := "http://139.162.9.169:8339/monitor"
	body, err := get(url)
	if err != nil {
		return nil, err
	}

	report := &Report{}
	err = json.Unmarshal(body, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func main() {
	filename := "keylist.json"
	keyListFromFile := common.KeyList{}
	jsonFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal([]byte(byteValue), &keyListFromFile)
	if err != nil {
		log.Fatal(err)
	}

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

func found(miningKey string, peers []struct{ Pubkey string }) bool {
	for _, p := range peers {
		if miningKey == p.Pubkey {
			return true
		}
	}
	return false
}
