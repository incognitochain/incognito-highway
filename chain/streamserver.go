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
	_, req.Heights = capBlocksPerRequest(req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], req.Heights)
	g := NewBlkGetter(req, nil)
	blkRecv := g.Get(ctx, s)
	logger.Infof("[stream] listen gblkRecv Start")
	for blk := range blkRecv {
		if len(blk.Data) == 0 {
			return nil
		}
		if err := ss.Send(&proto.BlockData{Data: blk.Data}); err != nil {
			logger.Infof("[stream] Trying send to client but received error %v, return and cancel context", err)
			return err
		}
	}
	logger.Infof("[stream] listen gblkRecv End")
	return nil
}

func (s *Server) StreamBlockByHash(
	req *proto.BlockByHashRequest,
	ss proto.HighwayService_StreamBlockByHashServer,
) error {
	if req.GetCallDepth() > common.MaxCallDepth {
		return errors.Errorf("reach max calldepth %v ", req)
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Infof("Receive StreamBlockByHash request, type = %s, hashes = %v %v", req.GetType().String(), req.GetHashes()[0], req.GetHashes()[len(req.GetHashes())-1])

	g := NewBlkGetter(nil, req)
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
