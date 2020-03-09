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
	defer close(blkChan)
	logger.Infof("[stream] Server call Client: Start stream request %v", req)
	sc, _, err := c.getClientWithBlock(ctx, int(req.GetFrom()), req.GetHeights()[len(req.GetHeights())-1])
	if sc == nil {
		logger.Warnf("[stream] Client is nil!")
		blkChan <- common.ExpectedBlk{
			Height: req.GetHeights()[0],
			Data:   []byte{},
		}
		return nil
	}
	stream, err := sc.StreamBlockByHeight(ctx, req.(*proto.BlockByHeightRequest))
	if err != nil {
		logger.Infof("[stream] Server call Client return error %v", err)
		return err
	}
	defer stream.CloseSend()
	logger.Infof("[stream] Server call Client: OK, return stream %v", stream)
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
