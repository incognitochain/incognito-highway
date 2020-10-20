package simulateutils

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/wire"
)

func UnmarshalShardBlock(blockString []byte) (*blockchain.ShardBlock, error) {
	var shardBlk blockchain.ShardBlock
	err := json.Unmarshal(blockString, &shardBlk)
	if err != nil {
		return nil, err
	}
	return &shardBlk, nil
}

func CheckIfExistTxIDInBlk(msg *wire.MessageBFT, targetTxID string) bool {
	if msg.ChainKey == "beacon" {
		return false
	}
	block, err := UnmarshalShardBlock(msg.Content)
	if err != nil {
		return false
	}
	for _, tx := range block.Body.Transactions {
		if tx.Hash().String() == targetTxID {
			return true
		}
	}
	return false
}
