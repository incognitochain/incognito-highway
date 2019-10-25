package common

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common/base58"
)

// func GetListMsgForProxySubOfCommittee(committeeID byte) []string {

// }

func (pubKey *CommitteePublicKey) FromString(keyString string) error {
	keyBytes, ver, err := base58.Base58Check{}.Decode(keyString)
	if (ver != 0x00) || (err != nil) {
		return errors.New("Decode key error")
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

func GetCommitteeIDOfValidator(validator string) int {
	// TODO(0xbunyip): remove hardcode here (testing)
	st1 := "1Eh5UZjDSUsdofnToKSELNz24rUfKgdsf3JubsetDQYscHHWERnQcVUxee2SsN4W1D8kdBwGdxmHM8FKUJieTEeowQBi4XJmAYhMc7Kmu19QeovmbLg9uvqch7Mh9ZDZ2vLBBi2iLzy6UQc4Bn5rRS9VdQC4jzrUrCFQqgjoL7QHnDascP7jNBB4iucn55gJgPVGsBAP3BuZcTkphpriLpZYk3f6hdWwb2P8EG5WpY8xWzSHGTxgoNZVgJTXVA78iK8d4ZJpVt3zHqTN9TXjpsJ7rRtHZrvK588meLUkk55kdbQnpm4Xg8fYWJ8jV7BiMmDGFWdYqhKiqvMNPgZ7MPNUbwiv6ugtABqGHJWcoYomQXAaaLz5EQ"
	if validator == st1 {
		return 0
	}

	if id, exist := CommitteeGenesis[validator]; exist {
		return int(id)
	} else {
		validatorKey := new(CommitteePublicKey)
		validatorKey.FromString(validator)
		validatorMiningPK, err := validatorKey.MiningPublicKey()
		if err != nil {
			return -1
		}
		if fullKey, ok := MiningKeyByCommitteeKey[validatorMiningPK]; ok {
			if fullKeyID, isExist := CommitteeGenesis[fullKey]; isExist {
				return int(fullKeyID)
			}
		}
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
