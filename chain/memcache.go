package chain

import (
	context "context"
	"fmt"

	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

type MemCache struct {
	cacher *ristretto.Cache
}

func NewMemCache() (*MemCache, error) {
	// TODO(@0xbunyip): move constants
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000000,
		MaxCost:     8 * 2 << 30, // 8 GiB
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		return nil, err
	}

	return &MemCache{
		cacher: cache,
	}, nil
}

func getKeyByHeight(req RequestByHeight, h uint64) string {
	return fmt.Sprintf("byheight-%d-%d-%d", req.fromShard, req.toShard, h)
}

func (cache *MemCache) GetBlockByHeight(_ context.Context, req RequestByHeight, heights []uint64) ([][]byte, error) {
	// TODO(@0xbunyip): add epsilon-greedy here
	blocks := make([][]byte, len(heights))
	for i, h := range heights {
		key := getKeyByHeight(req, h)
		if b, ok := cache.cacher.Get(key); ok {
			if block, ok := b.([]byte); ok {
				blocks[i] = block
			}
		}
	}
	return blocks, nil
}

func (cache *MemCache) SetBlockByHeight(_ context.Context, req RequestByHeight, heights []uint64, blocks [][]byte) error {
	if len(heights) != len(blocks) {
		return errors.Errorf("invalid blocks to cache: len(heights) = %d, len(blocks) = %d", len(heights), len(blocks))
	}

	for i, h := range heights {
		block := blocks[i]
		if block == nil || len(block) == 0 {
			continue
		}

		key := getKeyByHeight(req, h)
		cost := int64(len(block)) // Cost is the size of the block ==> limit maximum memory used by the cache
		cache.cacher.Set(key, block, cost)
	}
	return nil
}

func (cache *MemCache) GetBlockByHash(_ context.Context, req RequestByHash, hashes [][]byte) ([][]byte, error) {
	blocks := make([][]byte, len(hashes)) // Not supported
	return blocks, nil
}
