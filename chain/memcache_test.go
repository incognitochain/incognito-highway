package chain_test

import (
	context "context"
	"highway/chain"
	"highway/common"
	"highway/proto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBlockByHeight(t *testing.T) {
	cacher := &Cacher{}
	cacher.On("Get", mock.Anything).Return([]byte{}, false)
	cache := chain.NewMemCache(cacher)

	req := &proto.BlockByHeightRequest{}
	req.Heights = []uint64{3, 4, 5, 6}
	req.Specific = true
	blkChan := make(chan common.ExpectedBlkByHeight, 10)
	err := cache.StreamBlkByHeight(context.Background(), req, blkChan)

	if assert.Nil(t, err) {
		expectedBlocks := [][]byte{[]byte{}, []byte{}, []byte{}, []byte{}}
		blks := [][]byte{}
		for blk := range blkChan {
			blks = append(blks, blk.Data)
		}
		assert.Equal(t, expectedBlocks, blks)
	}
}

func TestSetSingleBlockByHeight(t *testing.T) {
	cacher := &Cacher{}
	sets := [][]byte{}
	cacher.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(true).Run(func(args mock.Arguments) {
		sets = append(sets, args.Get(1).([]byte))
	})
	cache := chain.NewMemCache(cacher)

	req := &proto.BlockByHeightRequest{
		From: int32(1),
		To:   int32(2),
	}
	blk := common.ExpectedBlkByHeight{
		Height: 3,
		Data:   []byte{4},
	}
	err := cache.SetSingleBlockByHeight(context.Background(), req, blk)

	if assert.Nil(t, err) {
		expectedSets := [][]byte{[]byte{4}}
		assert.Equal(t, expectedSets, sets)
	}
}
