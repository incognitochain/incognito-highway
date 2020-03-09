package chain

import (
	context "context"
	"highway/common"
	"highway/proto"

	"github.com/pkg/errors"
)

func (s *Server) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	ss proto.HighwayService_StreamBlockByHeightServer,
) error {
	if req.GetCallDepth() > common.MaxCallDepth {
		return errors.Errorf("reach max calldepth %v ", req)
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Infof("Receive StreamBlockByHeight request, type = %s, heights = %v %v", req.GetType().String(), req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1])

	g := NewBlkGetter(req)
	blkRecv := g.Get(ctx, s)
	logger.Infof("[stream] listen gblkRecv Start")
	for blk := range blkRecv {
		if len(blk.Data) == 0 {
			return errors.New("close")
		}
		logger.Infof("[stream] Received block from channel, send to client")
		if err := ss.Send(&proto.BlockData{Data: blk.Data}); err != nil {
			logger.Infof("[stream] Trying send to client but received error %v, return and cancel context", err)
			logger.Infof("[stream] listen gblkRecv End")
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
	logger.Infof("[stream] listen gblkRecv End")
	return nil
}
