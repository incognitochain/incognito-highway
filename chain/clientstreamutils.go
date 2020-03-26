package chain

import (
	context "context"
	"highway/common"
	"highway/proto"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

func (c *Client) StreamBlkByHeight(
	ctx context.Context,
	req RequestBlockByHeight,
	blkChan chan common.ExpectedBlkByHeight,
) error {
	logger := Logger(ctx)
	var stream proto.HighwayService_StreamBlockByHeightClient
	defer close(blkChan)
	logger.Infof("[stream] Server call Client: Start stream request, type %s, shard %d -> %d, #heights %d", req.GetType().String(), req.GetFrom(), req.GetTo(), len(req.GetHeights()))
	sc, _, err := c.getClientWithBlock(ctx, int(req.GetFrom()), req.GetHeights()[len(req.GetHeights())-1])
	if (err != nil) || (sc == nil) {
		err = errors.Errorf("[stream] getClientWithBlock return error %v, sc return %v", err, sc)
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
				blkChan <- common.ExpectedBlkByHeight{
					Height: blkHeight,
					Data:   blkData.GetData(),
				}
				continue
			} else {
				logger.Infof("[stream] Received err %v %v", stream, err)
			}
		}
		blkChan <- common.ExpectedBlkByHeight{
			Height: blkHeight,
			Data:   []byte{},
		}
	}
	return nil
}

func (c *Client) StreamBlkByHeightv2(
	ctx context.Context,
	req RequestBlockByHeight,
	blkChan chan common.ExpectedBlk,
) error {
	logger := Logger(ctx)
	var stream proto.HighwayService_StreamBlockByHeightClient
	defer close(blkChan)
	logger.Infof("[stream] Server call Client: Start stream request from peer %v, type %s, shard %d -> %d, #heights %d", req.GetSyncFromPeer(), req.GetType().String(), req.GetFrom(), req.GetTo(), len(req.GetHeights()))

	var (
		sc  proto.HighwayServiceClient
		err error
		pID peer.ID
	)
	if len(req.GetSyncFromPeer()) > 0 {
		pID, err = peer.IDFromString(req.GetSyncFromPeer())
		if err == nil {
			sc, err = c.cc.GetServiceClient(pID)
		}
	} else {
		sc, _, err = c.getClientWithBlock(ctx, int(req.GetFrom()), req.GetHeights()[len(req.GetHeights())-1])
	}

	if (err != nil) || (sc == nil) {
		err = errors.Errorf("[stream] getClientWithBlock return error %v, sc return %v", err, sc)
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
				blkChan <- common.ExpectedBlk{
					Height: blkHeight,
					Hash:   []byte{},
					Data:   blkData.GetData(),
				}
				continue
			} else {
				logger.Infof("[stream] Received err %v %v", stream, err)
			}
		}
		blkChan <- common.ExpectedBlk{
			Height: blkHeight,
			Hash:   []byte{},
			Data:   []byte{},
		}
	}
	return nil
}

func (c *Client) StreamBlkByHash(
	ctx context.Context,
	req RequestBlockByHash,
	blkChan chan common.ExpectedBlk,
) error {
	logger := Logger(ctx)
	var stream proto.HighwayService_StreamBlockByHashClient
	defer close(blkChan)
	logger.Infof("[stream] Server call Client: Start stream request, type %s, shard %d -> %d, #hash %d", req.GetType().String(), req.GetFrom(), req.GetTo(), len(req.GetHashes()))
	var (
		sc  proto.HighwayServiceClient
		err error
		pID peer.ID
	)
	if len(req.GetSyncFromPeer()) > 0 {
		pID, err = peer.IDFromString(req.GetSyncFromPeer())
		if err == nil {
			sc, err = c.cc.GetServiceClient(pID)
		}
	}
	if err != nil {
		sc, _, err = c.getClientWithHashes(int(req.GetFrom()), req.GetHashes())
	}

	if (err != nil) || (sc == nil) {
		err = errors.Errorf("[stream] getClientWithBlock return error %v, sc return %v", err, sc)
		logger.Errorf("[stream] getClientWithBlock return error %v", err)
	} else {
		nreq, ok := req.(*proto.BlockByHashRequest)
		if !ok {
			err = errors.Errorf("Invalid Request %v", req)
		} else {
			nreq.CallDepth++
			stream, err = sc.StreamBlockByHash(ctx, nreq)
			if err != nil {
				logger.Errorf("[stream] Server call Client return error %v", err)
			} else {
				logger.Infof("[stream] Server call Client: OK, return stream %v", stream)
				defer stream.CloseSend()
				defer func(stream proto.HighwayService_StreamBlockByHashClient) {
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
	hashes := req.GetHashes()
	blkData := new(proto.BlockData)
	for _, blkHash := range hashes {
		if err == nil {
			blkData, err = stream.Recv()
			if err == nil {
				blkChan <- common.ExpectedBlk{
					Height: 0,
					Hash:   blkHash,
					Data:   blkData.GetData(),
				}
				continue
			} else {
				logger.Infof("[stream] Received err %v %v", stream, err)
			}
		}
		blkChan <- common.ExpectedBlk{
			Height: 0,
			Hash:   blkHash,
			Data:   []byte{},
		}
	}
	return nil
}
