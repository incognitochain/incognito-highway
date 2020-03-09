package chain

import (
	context "context"
	"highway/common"
	"highway/proto"
)

func (c *Client) StreamBlkByHeight(
	ctx context.Context,
	req RequestBlockByHeight,
	blkChan chan common.ExpectedBlk,
) error {
	logger := Logger(ctx)
	var stream proto.HighwayService_StreamBlockByHeightClient
	defer close(blkChan)
	logger.Infof("[stream] Server call Client: Start stream request %v", req)
	sc, _, err := c.getClientWithBlock(ctx, int(req.GetFrom()), req.GetHeights()[len(req.GetHeights())-1])
	if err != nil {
		logger.Errorf("[stream] getClientWithBlock return error %v", err)
	} else {
		stream, err = sc.StreamBlockByHeight(ctx, req.(*proto.BlockByHeightRequest))
		if err != nil {
			logger.Errorf("[stream] Server call Client return error %v", err)
		} else {
			logger.Infof("[stream] Server call Client: OK, return stream %v", stream)
			defer stream.CloseSend()
		}
	}
	heights := req.GetHeights()
	blkHeight := heights[0] - 1
	idx := 0
	blkData := new(proto.BlockData)
	for blkHeight < heights[len(heights)-1] {
		if req.GetSpecific() {
			blkHeight = heights[idx]
			idx++
		} else {
			blkHeight++
		}
		if err == nil {
			blkData, err = stream.Recv()
			if err == nil {
				logger.Infof("[stream] Received block %v", blkHeight)
				blkChan <- common.ExpectedBlk{
					Height: blkHeight,
					Data:   blkData.GetData(),
				}
				continue
			} else {
				logger.Infof("[stream] Received err %v %v", stream, err)
			}
		}
		blkChan <- common.ExpectedBlk{
			Height: blkHeight,
			Data:   []byte{},
		}
	}
	return nil
}
