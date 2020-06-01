package chain

import (
	context "context"
	"highway/common"
	"highway/proto"
)

type BlkGetter struct {
	waiting   map[uint64][]byte
	newBlk    chan common.ExpectedBlk
	newHeight uint64
	idx       int
	blkRecv   chan common.ExpectedBlk
	req       *proto.BlockByHeightRequest
}

func NewBlkGetter(req *proto.BlockByHeightRequest) *BlkGetter {
	g := &BlkGetter{}
	g.waiting = map[uint64][]byte{}
	g.newBlk = make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
	g.idx = 0
	g.req = req
	g.newHeight = g.req.Heights[0] - 1
	g.blkRecv = make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
	g.updateNewHeight()
	return g
}

func (g *BlkGetter) Get(ctx context.Context, s *Server) chan common.ExpectedBlk {
	go g.CallForBlocks(ctx, s.Providers)
	go g.listenCommingBlk(ctx)
	return g.blkRecv
}

func (g *BlkGetter) checkWaitingBlk() bool {
	for {
		if (g.newHeight == 0) || len(g.waiting) == 0 {
			return false
		}
		if data, ok := g.waiting[g.newHeight]; ok {
			g.blkRecv <- common.ExpectedBlk{
				Height: g.newHeight,
				Data:   data,
			}
			delete(g.waiting, g.newHeight)
			g.updateNewHeight()
		} else {
			break
		}
	}
	return false
}

func (g *BlkGetter) listenCommingBlk(ctx context.Context) {
	defer close(g.blkRecv)
	for blk := range g.newBlk {
		if blk.Height < g.newHeight {
			continue
		}
		if blk.Height == g.newHeight {
			g.blkRecv <- blk
			g.updateNewHeight()
		} else {
			g.waiting[blk.Height] = blk.Data
		}
		g.checkWaitingBlk()
	}
	g.checkWaitingBlk()
}

func (g *BlkGetter) updateNewHeight() {
	if g.newHeight == g.req.Heights[len(g.req.Heights)-1] {
		g.newHeight = 0
		return
	}
	if g.req.GetSpecific() {
		g.newHeight = g.req.Heights[g.idx]
		g.idx++
	} else {
		g.newHeight++
	}
}

func (g *BlkGetter) handleBlkRecv(
	ctx context.Context,
	req *proto.BlockByHeightRequest,
	ch chan common.ExpectedBlk,
	providers []Provider,
) []uint64 {
	logger := Logger(ctx)
	missing := []uint64{}
	for blk := range ch {
		if len(blk.Data) == 0 {
			missing = append(missing, blk.Height)
		} else {
			g.newBlk <- blk
			go func(providers []Provider, blk common.ExpectedBlk) {
				for _, p := range providers {
					err := p.SetSingleBlockByHeight(ctx, req, blk)
					if err != nil {
						logger.Errorf("[stream] Caching return error %v", err)
					}
				}
			}(providers, blk)
		}
	}
	return missing
}

func newReq(
	oldReq *proto.BlockByHeightRequest,
	missing []uint64,
) *proto.BlockByHeightRequest {
	if len(missing) == 0 {
		return nil
	}
	return &proto.BlockByHeightRequest{
		Type:      oldReq.Type,
		Specific:  true,
		Heights:   missing,
		From:      oldReq.From,
		To:        oldReq.To,
		CallDepth: oldReq.CallDepth,
		UUID:      oldReq.UUID,
	}
}

func (g *BlkGetter) CallForBlocks(
	ctx context.Context,
	providers []Provider,
) error {
	logger := Logger(ctx)
	nreq := g.req
	logger.Infof("[stream] calling provider for req")
	for i, p := range providers {
		if nreq == nil {
			break
		}
		blkCh := make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
		go p.StreamBlkByHeight(ctx, nreq, blkCh)
		missing := g.handleBlkRecv(ctx, nreq, blkCh, providers[:i])
		logger.Infof("[stream] Provider %v return %v block", i, len(nreq.GetHeights())-len(missing))
		nreq = newReq(nreq, missing)
	}
	close(g.newBlk)
	return nil
}
