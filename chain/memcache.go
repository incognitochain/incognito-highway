package chain

import (
	context "context"
	"fmt"
	"highway/common"
	"math/rand"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

type MemCache struct {
	cacher Cacher
}

type Cacher interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
}

func NewMemCache(cacher Cacher) *MemCache {
	return &MemCache{cacher: cacher}
}

func NewRistrettoMemCache() (*MemCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: common.CacheNumCounters,
		MaxCost:     common.CacheMaxCost,
		BufferItems: common.CacheBufferItems,
		Metrics:     true,
	})
	if err != nil {
		return nil, err
	}
	return NewMemCache(cache), nil
}

func getKeyByHeight(req GetBlockByHeightRequest, h uint64) string {
	return fmt.Sprintf("byheight-%d-%d-%d", req.GetFrom(), req.GetTo(), h)
}

func (cache *MemCache) SetBlockByHeight(
	ctx context.Context,
	req GetBlockByHeightRequest,
	heights []uint64,
	blocks [][]byte,
) error {
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

func (cache *MemCache) SetSingleBlockByHeight(
	ctx context.Context,
	req RequestBlockByHeight,
	blk common.ExpectedBlkByHeight,
) error {
	// logger := Logger(ctx)
	if len(blk.Data) == 0 {
		return errors.Errorf("Block height %v has empty data", blk.Height)
	}
	// logger.Debugf("Caching block %s, height %d, shard %d -> %d, len = %d", req.GetType().String(), blk.Height, req.GetFrom(), req.GetTo(), len(blk.Data))

	key := keyByHeight(req, blk.Height)
	cost := int64(len(blk.Data)) // Cost is the size of the block ==> limit maximum memory used by the cache
	cache.cacher.Set(key, blk.Data, cost)
	return nil
}

func (cache *MemCache) SetSingleBlockByHeightv2(
	ctx context.Context,
	req RequestBlockByHeight,
	blk common.ExpectedBlk,
) error {
	// logger := Logger(ctx)
	if len(blk.Data) == 0 {
		return errors.Errorf("Block height %v has empty data", blk.Height)
	}
	// logger.Debugf("Caching block %s, height %d, shard %d -> %d, len = %d", req.GetType().String(), blk.Height, req.GetFrom(), req.GetTo(), len(blk.Data))

	key := keyByHeight(req, blk.Height)
	cost := int64(len(blk.Data)) // Cost is the size of the block ==> limit maximum memory used by the cache
	cache.cacher.Set(key, blk.Data, cost)
	return nil
}

func (cache *MemCache) SetSingleBlockByHash(
	ctx context.Context,
	req RequestBlockByHash,
	blk common.ExpectedBlk,
) error {
	// logger := Logger(ctx)
	if len(blk.Data) == 0 {
		return errors.Errorf("Block height %v has empty data", blk.Height)
	}
	// logger.Debugf("Caching block %s, height %d, shard %d -> %d, len = %d", req.GetType().String(), blk.Height, req.GetFrom(), req.GetTo(), len(blk.Data))

	key := keyByHash(req, blk.Hash)
	cost := int64(len(blk.Data)) // Cost is the size of the block ==> limit maximum memory used by the cache
	cache.cacher.Set(key, blk.Data, cost)
	return nil
}

func (cache *MemCache) GetBlockByHash(_ context.Context, req GetBlockByHashRequest, hashes [][]byte) ([][]byte, error) {
	blocks := make([][]byte, len(hashes)) // Not supported
	return blocks, nil
}

func (cache *MemCache) Metrics() map[string]interface{} {
	metric := map[string]interface{}{}
	if rcache, ok := cache.cacher.(*ristretto.Cache); ok {
		metric = map[string]interface{}{
			"ratio":        rcache.Metrics.Ratio(),
			"cost_added":   rcache.Metrics.CostAdded(),
			"cost_evicted": rcache.Metrics.CostEvicted(),
			"gets_kept":    rcache.Metrics.GetsKept(),
			"keys_added":   rcache.Metrics.KeysAdded(),
			"keys_evicted": rcache.Metrics.KeysEvicted(),
		}
	}
	return metric
}

func keyByHeight(req RequestBlockByHeight, h uint64) string {
	return fmt.Sprintf("byheight-%d-%d-%d", req.GetFrom(), req.GetTo(), h)
}

func keyByHash(req RequestBlockByHash, h []byte) string {
	return fmt.Sprintf("byhash-%d-%d-%v", req.GetFrom(), req.GetTo(), h)
}

func (cache *MemCache) StreamBlkByHeight(
	_ context.Context,
	req RequestBlockByHeight,
	blkChan chan common.ExpectedBlk,
) error {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	heights := req.GetHeights()
	blkHeight := heights[0] - 1
	idx := 0
	for blkHeight < heights[len(heights)-1] {
		if req.GetSpecific() {
			blkHeight = heights[idx]
			idx++
		} else {
			blkHeight++
		}
		key := keyByHeight(req, blkHeight)
		drop := r1.Intn(100)
		if drop <= common.PercentGetFromCache {
			if b, ok := cache.cacher.Get(key); ok {
				if block, ok := b.([]byte); ok {
					blkChan <- common.ExpectedBlk{
						Height: blkHeight,
						Hash:   []byte{},
						Data:   block,
					}
					continue
				}
			}
		}
		blkChan <- common.ExpectedBlk{
			Height: blkHeight,
			Hash:   []byte{},
			Data:   []byte{},
		}
	}
	close(blkChan)
	return nil
}

func (cache *MemCache) StreamBlkByHash(
	_ context.Context,
	req RequestBlockByHash,
	blkChan chan common.ExpectedBlk,
) error {
	hashes := req.GetHashes()
	for _, blkHash := range hashes {
		key := keyByHash(req, blkHash)
		if b, ok := cache.cacher.Get(key); ok {
			if block, ok := b.([]byte); ok {
				blkChan <- common.ExpectedBlk{
					Height: 0,
					Hash:   blkHash,
					Data:   block,
				}
				continue
			}
		}
		blkChan <- common.ExpectedBlk{
			Height: 0,
			Hash:   blkHash,
			Data:   []byte{},
		}
	}
	close(blkChan)
	return nil
}
