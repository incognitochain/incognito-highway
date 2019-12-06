package main

import (
	"encoding/json"
	"highway/common"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Peer struct {
	Pubkey string
	ID     string
}

type Report struct {
	Chain struct {
		Peers          map[int][]Peer            `json:"peers"`
		RequestPerPeer map[string]map[string]int `json:"request_per_peer"`
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
	url := "http://51.89.41.31:8080/monitor"
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

func loadKeylist(filename string) common.KeyList {
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
	return keyListFromFile
}

func found(miningKey string, peers []Peer) bool {
	for _, p := range peers {
		if miningKey == p.Pubkey {
			return true
		}
	}
	return false
}
