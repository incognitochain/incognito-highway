package chain

import (
	context "context"
	"highway/common"
	"highway/proto"
)

// type StreamServer struct {
// 	proto.UnimplementedHighwayStreamServiceServer
// 	blkgetter BlockGetter
// }

// type BlockGetter interface {
// 	GetBlocks(req *proto.BlockBeaconByHeightRequest) (chan interface{}, error)
// }

func (s *Server) StreamBlockBeaconByHeight(
	req *proto.GetBlockBeaconByHeightRequest,
	ss proto.HighwayService_StreamBlockBeaconByHeightServer,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	serviceClient, _, err := s.m.client.getClientWithBlock(ctx, int(common.BEACONID), req.GetToHeight())
	logger.Infof("[stream] Serve received request, start requesting another client: block by height: shard %v -> %v, heights = %v", req.GetFromHeight(), req.GetToHeight(), req.GetHeights())
	if err != nil {
		logger.Infof("[stream] No serviceClient with block, shardID = %v, heights = %v, err = %+v", req.GetFromHeight(), req.GetToHeight(), err)
		return err
	}
	blkRecv, err := GetBlocks(serviceClient, req)
	if err != nil {
		logger.Infof("[stream] Calling client but received error %v, return", err)
		return err
	}
	for blk := range blkRecv {
		logger.Infof("[stream] Received block from channel, send to client")
		if err := ss.Send(&proto.BlockData{Data: blk.([]byte)}); err != nil {
			logger.Infof("[stream] Trying send to client but received error %v, return", err)
			return err
		}
	}
	return nil
}
