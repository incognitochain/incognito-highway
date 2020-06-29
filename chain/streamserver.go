package chain

import (
	context "context"
	"highway/common"
	"highway/proto"

	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
)

func (s *Server) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	ss proto.HighwayService_StreamBlockByHeightServer,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	pClient, ok := peer.FromContext(ss.Context())
	pIP := "Can not get IP, so sorry"
	if ok {
		pIP = pClient.Addr.String()
	}
	logger.Infof("Receive StreamBlockByHeight request from IP: %v, type = %s - specific %v, heights = %v %v #%v", pIP, req.GetType().String(), req.Specific, req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1], len(req.GetHeights()))
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
