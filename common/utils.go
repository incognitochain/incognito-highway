package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func GetCommitteeIDOfValidator(validator string) int {
	if value, exist := CommitteeGenesis[validator]; exist {
		return int(value)
	}
	return -1
}

type Key struct {
	Payment         string `json:"PaymentAddress"`
	CommitteePubKey string `json:"CommitteePublicKey"`
}

type KeyList struct {
	Bc []Key         `json:"Beacon"`
	Sh map[int][]Key `json:"Shard"`
}

func InitGenesisCommitteeFromFile(filename string, numberOfShard, numberOfCandidate int) error {
	CommitteeGenesis = map[string]byte{}
	keyListFromFile := KeyList{}
	if filename != "" {
		jsonFile, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("Successfully Opened %v\n", filename)
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal([]byte(byteValue), &keyListFromFile)

	}

	for i := 0; i < numberOfCandidate; i++ {
		if i < len(keyListFromFile.Bc) {
			CommitteeGenesis[keyListFromFile.Bc[i].CommitteePubKey] = BEACONID
		}
	}
	for j := 0; j < numberOfShard; j++ {
		for i := 0; i < numberOfCandidate; i++ {
			if i < len(keyListFromFile.Sh[j]) {
				CommitteeGenesis[keyListFromFile.Sh[j][i].CommitteePubKey] = byte(j)
			}
		}
	}
	return nil
}
