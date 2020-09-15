package chain

// import (
// 	context "context"
// 	"highway/common"
// 	"highway/proto"
// )

// type BlkByHeightGetter struct {
// 	waiting   map[uint64][]byte
// 	newBlk    chan common.ExpectedBlkByHeight
// 	newHeight uint64
// 	idx       int
// 	blkRecv   chan common.ExpectedBlkByHeight
// 	req       *proto.BlockByHeightRequest
// }

// func NewBlkByHeightGetter(req *proto.BlockByHeightRequest) *BlkByHeightGetter {
// 	g := &BlkByHeightGetter{}
// 	g.waiting = map[uint64][]byte{}
// 	g.newBlk = make(chan common.ExpectedBlkByHeight, common.MaxBlocksPerRequest)
// 	g.idx = 0
// 	g.req = req
// 	g.newHeight = g.req.Heights[0] - 1
// 	g.blkRecv = make(chan common.ExpectedBlkByHeight, common.MaxBlocksPerRequest)
// 	g.updateNewHeight()
// 	return g
// }

// func (g *BlkByHeightGetter) Get(ctx context.Context, s *Server) chan common.ExpectedBlkByHeight {
// 	go g.CallForBlocksByHeight(ctx, s.Providers)
// 	go g.listenCommingBlk(ctx)
// 	return g.blkRecv
// }

// func (g *BlkByHeightGetter) checkWaitingBlk() bool {
// 	for {
// 		if (g.newHeight == 0) || len(g.waiting) == 0 {
// 			return false
// 		}
// 		if data, ok := g.waiting[g.newHeight]; ok {
// 			g.blkRecv <- common.ExpectedBlkByHeight{
// 				Height: g.newHeight,
// 				Data:   data,
// 			}
// 			delete(g.waiting, g.newHeight)
// 			g.updateNewHeight()
// 		} else {
// 			break
// 		}
// 	}
// 	return false
// }

// func (g *BlkByHeightGetter) listenCommingBlk(ctx context.Context) {
// 	defer close(g.blkRecv)
// 	for blk := range g.newBlk {
// 		if blk.Height < g.newHeight {
// 			continue
// 		}
// 		if blk.Height == g.newHeight {
// 			g.blkRecv <- blk
// 			g.updateNewHeight()
// 		} else {
// 			g.waiting[blk.Height] = blk.Data
// 		}
// 		g.checkWaitingBlk()
// 	}
// 	g.checkWaitingBlk()
// }

// func (g *BlkByHeightGetter) updateNewHeight() {
// 	if g.newHeight == g.req.Heights[len(g.req.Heights)-1] {
// 		g.newHeight = 0
// 		return
// 	}
// 	if g.req.GetSpecific() {
// 		g.newHeight = g.req.Heights[g.idx]
// 		g.idx++
// 	} else {
// 		g.newHeight++
// 	}
// }

// func (g *BlkByHeightGetter) handleBlkRecv(
// 	ctx context.Context,
// 	req *proto.BlockByHeightRequest,
// 	ch chan common.ExpectedBlkByHeight,
// 	providers []Provider,
// ) []uint64 {
// 	logger := Logger(ctx)
// 	missing := []uint64{}
// 	for blk := range ch {
// 		if len(blk.Data) == 0 {
// 			missing = append(missing, blk.Height)
// 		} else {
// 			g.newBlk <- blk
// 			go func(providers []Provider, blk common.ExpectedBlkByHeight) {
// 				for _, p := range providers {
// 					err := p.SetSingleBlockByHeight(ctx, req, blk)
// 					if err != nil {
// 						logger.Errorf("[stream] Caching return error %v", err)
// 					}
// 				}
// 			}(providers, blk)
// 		}
// 	}
// 	return missing
// }

// func newReqByHeight(
// 	oldReq *proto.BlockByHeightRequest,
// 	missing []uint64,
// ) *proto.BlockByHeightRequest {
// 	if len(missing) == 0 {
// 		return nil
// 	}
// 	return &proto.BlockByHeightRequest{
// 		Type:         oldReq.Type,
// 		Specific:     true,
// 		Heights:      missing,
// 		From:         oldReq.From,
// 		To:           oldReq.To,
// 		CallDepth:    oldReq.CallDepth,
// 		SyncFromPeer: oldReq.SyncFromPeer,
// 		UUID:         oldReq.UUID,
// 	}
// }

// func (g *BlkByHeightGetter) CallForBlocksByHeight(
// 	ctx context.Context,
// 	providers []Provider,
// ) error {
// 	logger := Logger(ctx)
// 	nreq := g.req
// 	logger.Infof("[stream] calling provider for req")
// 	for i, p := range providers {
// 		if nreq == nil {
// 			break
// 		}
// 		blkCh := make(chan common.ExpectedBlkByHeight, common.MaxBlocksPerRequest)
// 		go p.StreamBlkByHeight(ctx, nreq, blkCh)
// 		missing := g.handleBlkRecv(ctx, nreq, blkCh, providers[:i])
// 		logger.Infof("[stream] Provider %v return %v block", i, getReqNumBlks(nreq)-len(missing))
// 		nreq = newReqByHeight(nreq, missing)
// 	}
// 	close(g.newBlk)
// 	return nil
// }
