package chain

import (
	context "context"
	"highway/common"
	"highway/proto"

	"github.com/pkg/errors"
)

func (c *Client) StreamBlkByHeight(
	ctx context.Context,
	req RequestBlockByHeight,
	blkChan chan common.ExpectedBlk,
) error {
	logger := Logger(ctx)
	logger.Infof("[stream] StreamBlkByHeight Start")
	var stream proto.HighwayService_StreamBlockByHeightClient
	defer close(blkChan)
	logger.Infof("[stream] Server call Client: Start stream request %v", req)
	sc, _, err := c.getClientWithBlock(ctx, int(req.GetFrom()), req.GetHeights()[len(req.GetHeights())-1])
	if err != nil {
		logger.Errorf("[stream] getClientWithBlock return error %v", err)
	} else {
		nreq, ok := req.(*proto.BlockByHeightRequest)
		if !ok {
			err = errors.Errorf("Invalid Request %v", req)
		} else {
			nreq.CallDepth++
			stream, err = sc.StreamBlockByHeight(ctx, nreq)
			if err != nil {
				logger.Errorf("[stream] Server call Client return error %v", err)
			} else {
				logger.Infof("[stream] Server call Client: OK, return stream %v", stream)
				defer stream.CloseSend()
				defer func(stream proto.HighwayService_StreamBlockByHeightClient) {
					for {
						_, errStream := stream.Recv()
						if errStream != nil {
							break
						}
					}
				}(stream)
			}
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
	logger.Infof("[stream] StreamBlkByHeight End")
	return nil
}
