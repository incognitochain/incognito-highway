package chain

import (
	context "context"
	"errors"
	"highway/common"
	"highway/proto"
)

func (s *Server) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	ss proto.HighwayService_StreamBlockByHeightServer,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Infof("Receive StreamBlockByHeight request, type = %s, heights = %v %v", req.GetType().String(), req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1])

	g := NewBlkGetter(req)
	defer g.Cancel()
	blkRecv := g.Get(ctx, s)
	for blk := range blkRecv {
		if len(blk.Data) == 0 {
			return errors.New("close")
		}
		logger.Infof("[stream] Received block from channel, send to client")
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
