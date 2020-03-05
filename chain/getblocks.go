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
	g.newBlk = make(chan common.ExpectedBlk)
	g.idx = 0
	g.req = req
	g.newHeight = g.req.Heights[0] - 1
	g.blkRecv = make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
	g.updateNewHeight()
	return g
}

func (g *BlkGetter) Cancel() {
	for range g.blkRecv {
	}
}

func (g *BlkGetter) Get(ctx context.Context, s *Server) chan common.ExpectedBlk {
	go g.CallForBlocks(ctx, s.Providers)
	go g.listenCommingBlk(ctx)
	return g.blkRecv
}

func (g *BlkGetter) checkWaitingBlk() bool {
	if data, ok := g.waiting[g.newHeight]; ok {
		g.blkRecv <- common.ExpectedBlk{
			Height: g.newHeight,
			Data:   data,
		}
		delete(g.waiting, g.newHeight)
		g.updateNewHeight()
		return true
	}
	return false
}

func (g *BlkGetter) listenCommingBlk(ctx context.Context) {
	logger := Logger(ctx)
	defer close(g.blkRecv)
	for blk := range g.newBlk {
		logger.Infof("[stream] ListenComming received %v, wanted %v", blk.Height, g.newHeight)
		if blk.Height < g.newHeight {
			continue
		}
		if blk.Height == g.newHeight {
			g.blkRecv <- blk
			g.updateNewHeight()
		} else {
			g.waiting[blk.Height] = blk.Data
			g.checkWaitingBlk()
		}
	}
	for {
		if (g.newHeight == 0) || len(g.waiting) == 0 {
			return
		}
		ok := g.checkWaitingBlk()
		if !ok {
			return
		}
	}
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

func (g *BlkGetter) handleBlkRecv(ctx context.Context, req *proto.BlockByHeightRequest, ch chan common.ExpectedBlk) []uint64 {
	logger := Logger(ctx)
	missing := []uint64{}
	for blk := range ch {
		logger.Infof("[stream] handleBlkRecv received block %v", blk.Height)
		if len(blk.Data) == 0 {
			missing = append(missing, blk.Height)
		} else {
			g.newBlk <- blk
		}
	}
	if len(missing) != 0 {
		last := missing[len(missing)-1]
		if req.Specific {
			for i, height := range req.Heights {
				if height > last {
					missing = append(missing, req.Heights[i:]...)
					break
				}
			}
		} else {
			for height := last + 1; height <= req.Heights[len(req.Heights)-1]; height++ {
				missing = append(missing, height)
			}
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
	}
}

func (g *BlkGetter) CallForBlocks(
	ctx context.Context,
	providers []Provider,
) error {
	logger := Logger(ctx)
	nreq := g.req
	for _, p := range providers {
		if nreq == nil {
			break
		}
		blkCh := make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
		logger.Infof("[stream] calling provider for req")
		go p.StreamBlkByHeight(ctx, nreq, blkCh)
		missing := g.handleBlkRecv(ctx, nreq, blkCh)
		nreq = newReq(nreq, missing)
	}
	close(g.newBlk)
	return nil
}
