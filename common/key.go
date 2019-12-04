package common

import (
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type ProcessedKey string

// PreprocessKey receives either CommitteePubkey or MiningPubkey;
// for CommitteePubKey, it converts to MiningPubKey and returns;
// otherwise, keep the key as it is
func PreprocessKey(pubkey string) (ProcessedKey, error) {
	key := new(incognitokey.CommitteePublicKey)
	if err := key.FromString(pubkey); err != nil {
		// Couldn't parsed, check if it's mining key
		_, _, err := base58.DecodeCheck(pubkey)
		if err != nil {
			return ProcessedKey(""), errors.WithMessagef(err, "pubkey: %v", pubkey)
		}

		// base58 decoded successfully, must be mining key
		return ProcessedKey(pubkey), nil
	}

	// Parsed successfully, get miningkey
	return ProcessedKey(key.GetMiningKeyBase58("bls")), nil
}
