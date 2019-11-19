package common

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common/base58"
)

// func GetListMsgForProxySubOfCommittee(committeeID byte) []string {

// }

func (pubKey *CommitteePublicKey) FromString(keyString string) error {
	keyBytes, ver, err := base58.Base58Check{}.Decode(keyString)
	if (ver != 0x00) || (err != nil) {
		return errors.New(fmt.Sprintf("Decode key error %v -%v-", err, keyString))
	}
	err = json.Unmarshal(keyBytes, pubKey)
	if err != nil {
		return err
	}
	return nil
}

func (pubKey *CommitteePublicKey) ToBase58() (string, error) {
	result, err := json.Marshal(pubKey)
	if err != nil {
		return "", err
	}
	return base58.Base58Check{}.Encode(result, 0x00), nil
}

func (pubKey *CommitteePublicKey) MiningPublicKey() (string, error) {
	result, err := json.Marshal(pubKey.MiningPubKey)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

type Key struct {
	Payment         string `json:"PaymentAddress"`
	CommitteePubKey string `json:"CommitteePublicKey"`
}

type KeyList struct {
	Bc []Key         `json:"Beacon"`
	Sh map[int][]Key `json:"Shard"`
}

func HasValuesAt(
	slice []byte,
	value byte,
) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func HasStringAt(
	slice []string,
	value string,
) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}
