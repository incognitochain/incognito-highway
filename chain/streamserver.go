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
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Infof("Receive StreamBlockByHeight request, type = %s - specific %v, heights = %v %v #%v", req.GetType().String(), req.Specific, req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1], len(req.GetHeights()))
	if req.GetCallDepth() > common.MaxCallDepth {
		err := errors.Errorf("reach max calldepth %v ", req)
		logger.Error(err)
		return err
	}
	if err := proto.CheckReq(req); err != nil {
		logger.Error(err)
		return err
	}
	g := NewBlkGetter(req)
	blkRecv := g.Get(ctx, s)
	// logger.Infof("[stream] listen gblkRecv Start")
	for blk := range blkRecv {
		if len(blk.Data) == 0 {
			return nil
		}
		if err := ss.Send(&proto.BlockData{Data: blk.Data}); err != nil {
			logger.Infof("[stream] Trying send to client but received error %v, return and cancel context", err)
			return err
		}
	}
	// logger.Infof("[stream] listen gblkRecv End")
	return nil
}
