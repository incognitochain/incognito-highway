package chain

import (
	context "context"
	"highway/proto"
	"io"
)

func streamData(
	stream proto.HighwayService_StreamBlockByHeightClient,
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
		blkRecv <- blkData.GetData()
	}
}

func getBlocks(
	ctx context.Context,
	c proto.HighwayServiceClient,
	req *proto.BlockByHeightRequest,
) (
	chan interface{},
	error,
) {
	logger.Infof("[stream] Server call Client: Start stream request %v", req)
	stream, err := c.StreamBlockByHeight(ctx, req)
	if err != nil {
		logger.Infof("[stream] Server call Client return error %v", err)
		return nil, err
	}
	logger.Infof("[stream] Server call Client: OK, return stream %v", stream)
	blkRecv := make(chan interface{})
	go streamData(stream, blkRecv)
	return blkRecv, nil
}
