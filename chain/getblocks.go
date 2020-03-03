package chain

import (
	context "context"
	"highway/common"
	"highway/proto"
)

type BlkGetter struct {
	listenNextBlk []chan uint64
	newBlk        chan common.ExpectedBlk
	newHeight     uint64
	idx           int
	blkRecv       chan []byte
	req           *proto.BlockByHeightRequest
}

func NewBlkGetter(req *proto.BlockByHeightRequest) *BlkGetter {
	g := &BlkGetter{}
	g.listenNextBlk = []chan uint64{}
	g.newBlk = make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
	g.idx = 0
	g.req = req
	g.newHeight = g.req.Heights[0] - 1
	g.blkRecv = make(chan []byte)
	g.updateNewHeight()
	return g
}

func (g *BlkGetter) publishNewHeights(height uint64) {
	for _, lisner := range g.listenNextBlk {
		lisner <- height
	}
}

func (g *BlkGetter) Cancel() {
	g.newBlk <- common.ExpectedBlk{
		Height: 0,
		Data:   []byte{},
	}
}

func (g *BlkGetter) Get(ctx context.Context, s *Server) chan []byte {
	go g.CallForBlocks(ctx, s.Providers)
	go g.listenCommingBlk()
	return g.blkRecv
}

func (g *BlkGetter) listenCommingBlk() {
	for blk := range g.newBlk {
		logger.Infof("[stream] ListenComming received %v, wanted %v", blk.Height, g.newHeight)
		if blk.Height < g.newHeight {
			if blk.Height == 0 {
				g.publishNewHeights(0)
				return
			}
			continue
		}
		if blk.Height == g.newHeight {
			g.blkRecv <- blk.Data
			g.updateNewHeight()
		}
		g.publishNewHeights(g.newHeight)
	}
}

func (g *BlkGetter) updateNewHeight() {
	if g.newHeight == g.req.Heights[len(g.req.Heights)-1] {
		g.blkRecv <- []byte{}
		return
	}
	if g.req.GetSpecific() {
		g.newHeight = g.req.Heights[g.idx]
		g.idx++
	} else {
		g.newHeight++
	}
}

func (g *BlkGetter) handleBlkRecv(ctx context.Context, ch chan common.ExpectedBlk) []uint64 {
	listenNewHeight := make(chan uint64)
	waiting := map[uint64][]byte{}
	missing := []uint64{}
	for blk := range ch {
		logger.Infof("[stream] handleBlkRecv %v received block %v", ctx.Value("ID"), blk.Height)
		if len(blk.Data) == 0 {
			missing = append(missing, blk.Height)
		} else {
			waiting[blk.Height] = blk.Data
			g.newBlk <- blk
		}
	}
	g.listenNextBlk = append(g.listenNextBlk, listenNewHeight)
	go g.listenForReturn(listenNewHeight, waiting)
	return missing
}

func (g *BlkGetter) listenForReturn(listener chan uint64, data map[uint64][]byte) {
	for h := range listener {
		if h == 0 {
			close(listener)
			return
		}
		logger.Infof("[stream] Listener Received height %v", h)
		if blkData, ok := data[h]; ok {
			g.newBlk <- common.ExpectedBlk{
				Height: h,
				Data:   blkData,
			}
			delete(data, h)
		}
		if len(data) == 0 {
			return
		}
	}
}

func newReq(
	oldReq *proto.BlockByHeightRequest,
	missing []uint64,
) *proto.BlockByHeightRequest {
	nReq := &proto.BlockByHeightRequest{
		Type:      oldReq.Type,
		Specific:  true,
		Heights:   missing,
		From:      oldReq.From,
		To:        oldReq.To,
		CallDepth: oldReq.CallDepth,
	}
	return nReq
}

func (g *BlkGetter) CallForBlocks(
	ctx context.Context,
	providers []Provider,
) error {
	nreq := g.req
	for _, p := range providers {
		if nreq == nil {
			break
		}
		blkCh := make(chan common.ExpectedBlk, common.MaxBlocksPerRequest)
		go p.StreamBlkByHeight(ctx, nreq, blkCh)
		missing := g.handleBlkRecv(ctx, blkCh)
		nreq = newReq(nreq, missing)
	}
	return nil
}
