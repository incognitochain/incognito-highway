package chain

import (
	"bytes"
	context "context"
	"fmt"
	"highway/common"
	"highway/proto"
)

type BlkGetter struct {
	waiting     map[string][]byte
	newBlk      chan common.ExpectedBlk
	idx         int
	blkRecv     chan common.ExpectedBlk
	requestType byte
	reqByHeight *proto.BlockByHeightRequest
	reqByHash   *proto.BlockByHashRequest
}

func NewBlkGetter(
	reqByHeight *proto.BlockByHeightRequest,
	reqByHash *proto.BlockByHashRequest,
) *BlkGetter {
	g := &BlkGetter{}
	g.waiting = map[string][]byte{}
	g.newBlk = make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
	g.idx = 0
	g.reqByHeight = reqByHeight
	g.reqByHash = reqByHash
	if g.reqByHeight != nil {
		g.requestType = 0
	}
	if g.reqByHash != nil {
		g.requestType = 1
	}
	g.blkRecv = make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
	return g
}

func (g *BlkGetter) Get(ctx context.Context, s *Server) chan common.ExpectedBlk {
	go g.CallForBlocks(ctx, s.Providers)
	go g.listenCommingBlk(ctx)
	return g.blkRecv
}

func (g *BlkGetter) updateWantedBlock(
	currentHeight *uint64,
	currentHash *[]byte,
) {
	if g.requestType == 0 {
		if *currentHeight == g.reqByHeight.Heights[len(g.reqByHeight.Heights)-1] {
			*currentHeight = 0
			return
		}
		if g.reqByHeight.GetSpecific() {
			*currentHeight = g.reqByHeight.Heights[g.idx]
			g.idx++
		} else {
			if *currentHeight == 0 {
				*currentHeight = g.reqByHeight.Heights[0]
			} else {
				*currentHeight++
			}
		}
		return
	}
	if g.requestType == 1 {
		if bytes.Equal(*currentHash, g.reqByHash.Hashes[len(g.reqByHash.Hashes)-1]) {
			*currentHash = []byte{}
			return
		}
		*currentHash = g.reqByHash.Hashes[g.idx]
		g.idx++
		return
	}
}

func (g *BlkGetter) listenCommingBlk(ctx context.Context) {
	defer close(g.blkRecv)
	currentHeight := uint64(0)
	currentHash := []byte{}
	g.updateWantedBlock(&currentHeight, &currentHash)
	for blk := range g.newBlk {
		if g.requestType == 0 {
			if blk.Height < currentHeight {
				continue
			}
			if blk.Height == currentHeight {
				g.blkRecv <- blk
				g.updateWantedBlock(&currentHeight, &currentHash)
			} else {
				g.waiting[getKeyOfExpectedBlk(&blk)] = blk.Data
			}
			g.checkWaitingBlk(&currentHeight, &currentHash)
		}
		if g.requestType == 1 {
			if bytes.Equal(blk.Hash, currentHash) {
				g.blkRecv <- blk
				g.updateWantedBlock(&currentHeight, &currentHash)
			} else {
				g.waiting[getKeyOfExpectedBlk(&blk)] = blk.Data
			}
			g.checkWaitingBlk(&currentHeight, &currentHash)
		}
	}
	g.checkWaitingBlk(&currentHeight, &currentHash)
}

func getKeyOfExpectedBlk(blk *common.ExpectedBlk) string {
	return fmt.Sprintf("%v-%v", blk.Height, blk.Hash)
}

func (g *BlkGetter) checkWaitingBlk(
	currentHeight *uint64,
	currentHash *[]byte,
) bool {
	for {

		if ((g.requestType == 0) && (*currentHeight == 0)) ||
			((g.requestType == 1) && (bytes.Equal(*currentHash, []byte{0}))) ||
			len(g.waiting) == 0 {
			return false
		}
		keyWantedBlk := getKeyOfExpectedBlk(&common.ExpectedBlk{Height: *currentHeight, Hash: *currentHash})
		if data, ok := g.waiting[keyWantedBlk]; ok {
			g.blkRecv <- common.ExpectedBlk{
				Height: *currentHeight,
				Hash:   *currentHash,
				Data:   data,
			}
			delete(g.waiting, keyWantedBlk)
			g.updateWantedBlock(currentHeight, currentHash)
		} else {
			break
		}

	}
	return false
}

func (g *BlkGetter) handleBlkByHashRecv(
	ctx context.Context,
	req *proto.BlockByHashRequest,
	ch chan common.ExpectedBlk,
	providers []Provider,
) [][]byte {
	// logger := Logger(ctx)
	missing := [][]byte{}
	for blk := range ch {
		if len(blk.Data) == 0 {
			missing = append(missing, blk.Hash)
		} else {
			g.newBlk <- blk
			if req.Type != proto.BlkType_BlkS2B {
				go func(providers []Provider, blk common.ExpectedBlk) {
					for _, p := range providers {
						err := p.SetSingleBlockByHash(ctx, req, blk)
						if err != nil {
							logger.Errorf("[stream] Caching return error %v", err)
						}
					}
				}(providers, blk)
			}
		}
	}
	return missing
}

func newReqByHash(
	oldReq *proto.BlockByHashRequest,
	missing [][]byte,
) *proto.BlockByHashRequest {
	if len(missing) == 0 {
		return nil
	}
	return &proto.BlockByHashRequest{
		Type:         oldReq.Type,
		Hashes:       missing,
		From:         oldReq.From,
		To:           oldReq.To,
		CallDepth:    oldReq.CallDepth,
		UUID:         oldReq.UUID,
		SyncFromPeer: oldReq.SyncFromPeer,
	}
}

func newReqByHeight(
	oldReq *proto.BlockByHeightRequest,
	missing []uint64,
) *proto.BlockByHeightRequest {
	if len(missing) == 0 {
		return nil
	}
	return &proto.BlockByHeightRequest{
		Type:         oldReq.Type,
		Specific:     true,
		Heights:      missing,
		From:         oldReq.From,
		To:           oldReq.To,
		CallDepth:    oldReq.CallDepth,
		SyncFromPeer: oldReq.SyncFromPeer,
		UUID:         oldReq.UUID,
	}
}

func (g *BlkGetter) CallForBlocks(
	ctx context.Context,
	providers []Provider,
) error {
	logger := Logger(ctx)
	nreqByHash := g.reqByHash
	nreqByHeight := g.reqByHeight
	logger.Infof("[stream] calling provider for req")
	for i, p := range providers {
		blkCh := make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
		switch g.requestType {
		case 1:
			if nreqByHash == nil {
				break
			}
			if g.reqByHash.Type == proto.BlkType_BlkS2B {
				if i == 0 {
					continue
				}
			}
			go p.StreamBlkByHash(ctx, nreqByHash, blkCh)
			missing := g.handleBlkByHashRecv(ctx, nreqByHash, blkCh, providers[:i])
			logger.Infof("[stream] Provider %v return %v block", i, getReqNumHashes(nreqByHash)-len(missing))
			nreqByHash = newReqByHash(nreqByHash, missing)
			break
		case 0:
			if nreqByHeight == nil {
				break
			}
			if g.reqByHeight.Type == proto.BlkType_BlkS2B {
				if i == 0 {
					continue
				}
			}
			go p.StreamBlkByHeight(ctx, nreqByHeight, blkCh)
			missing := g.handleBlkByHeightRecv(ctx, nreqByHeight, blkCh, providers[:i])
			logger.Infof("[stream] Provider %v return %v block", i, getReqNumBlks(nreqByHeight)-len(missing))
			nreqByHeight = newReqByHeight(nreqByHeight, missing)
			break
		}
	}
	close(g.newBlk)
	return nil
}

func (g *BlkGetter) handleBlkByHeightRecv(
	ctx context.Context,
	req *proto.BlockByHeightRequest,
	ch chan common.ExpectedBlk,
	providers []Provider,
) []uint64 {
	// logger := Logger(ctx)
	missing := []uint64{}
	for blk := range ch {
		if len(blk.Data) == 0 {
			missing = append(missing, blk.Height)
		} else {
			g.newBlk <- blk
			if req.Type != proto.BlkType_BlkS2B {
				go func(providers []Provider, blk common.ExpectedBlk) {
					for _, p := range providers {
						err := p.SetSingleBlockByHeightv2(ctx, req, blk)
						if err != nil {
							logger.Errorf("[stream] Caching return error %v", err)
						}
					}
				}(providers, blk)
			}
		}
	}
	return missing
}

func getReqNumBlks(
	req *proto.BlockByHeightRequest,
) int {
	if req.Specific {
		return len(req.Heights)
	}
	return int(req.Heights[len(req.Heights)-1] - req.Heights[0] + 1)
}

func getReqNumHashes(
	req *proto.BlockByHashRequest,
) int {
	return len(req.Hashes)
}
