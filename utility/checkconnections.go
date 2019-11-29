package main

import (
	"encoding/json"
	"highway/common"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

	for cid, peers := range r.Chain.Peers {
		for _, pubkey := range peers {
			key := new(common.CommitteePublicKey)
			key.FromString(pubkey)
		}
	}
}
