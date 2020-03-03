package chain

import (
	context "context"
	"errors"
	"fmt"
	"highway/common"
	"highway/proto"
)

func (s *Server) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	ss proto.HighwayService_StreamBlockByHeightServer,
) error {
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), "ID", fmt.Sprintf("%v%v%v", req.GetType(), req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1])), common.MaxTimePerRequest)
	defer cancel()
	g := NewBlkGetter(req)
	blkRecv := g.Get(ctx, s)
	for blk := range blkRecv {
		if len(blk.Data) == 0 {
			return errors.New("close")
		}
		logger.Infof("[stream] Received block from channel %v, send to client", ctx.Value("ID"))
		if err := ss.Send(&proto.BlockData{Data: blk.Data}); err != nil {
			logger.Infof("[stream] Trying send to client but received error %v, return and cancel context", err)
			return err
		}
		go func(s *Server, blk common.ExpectedBlk) {
			for _, p := range s.Providers {
				err := p.SetSingleBlockByHeight(ctx, req, blk)
				if err != nil {
					logger.Errorf("[stream] Caching return error %v", err)
				}
			}
		}(s, blk)
	}
	return nil
}
