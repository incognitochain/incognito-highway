package chain

import (
	context "context"
	"highway/proto"
	"io"
)

func getStreamData(
	stream proto.HighwayService_StreamBlockBeaconByHeightClient,
	blkRecv chan interface{},
) {
	defer close(blkRecv)
	for {
		blkData, err := stream.Recv()
		if err == io.EOF {
			logger.Infof("[stream] Return io.EOF %v", stream)
			break
		}
		if err != nil {
			logger.Infof("[stream] Received err %v %v", stream, err)
			break
		}
		blkRecv <- blkData.GetData
	}
}

func GetBlocks(
	ctx context.Context,
	c proto.HighwayServiceClient,
	req *proto.GetBlockBeaconByHeightRequest,
) (
	chan interface{},
	error,
) {
	logger.Infof("[stream] Server call Client: Start stream request %v %v %v %v", req.GetSpecific(), req.GetFromHeight(), req.GetToHeight(), len(req.GetHeights()))
	stream, err := c.StreamBlockBeaconByHeight(ctx, req)
	if err != nil {
		logger.Infof("[stream] Server call Client return error %v", err)
		return nil, err
	}
	logger.Infof("[stream] Server call Client: OK, return stream %v", stream)
	blkRecv := make(chan interface{})
	go getStreamData(stream, blkRecv)
	return blkRecv, nil
}
