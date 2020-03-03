package chain

import (
	context "context"
	"errors"
	"fmt"
	"highway/common"
	"highway/proto"
)

func (s *Server) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	ss proto.HighwayService_StreamBlockByHeightServer,
) error {
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), "ID", fmt.Sprintf("%v%v%v", req.GetType(), req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1])), common.MaxTimePerRequest)
	defer cancel()
	g := NewBlkGetter(req)
	blkRecv := g.Get(ctx, s)
	defer g.Cancel()
	for blk := range blkRecv {
		if len(blk) == 0 {
			return errors.New("close")
		}
		logger.Infof("[stream] Received block from channel %v, send to client", ctx.Value("ID"))
		if err := ss.Send(&proto.BlockData{Data: blk}); err != nil {
			logger.Infof("[stream] Trying send to client but received error %v, return and cancel context", err)
			return err
		}
	}
	return nil
}

// logger.Infof("[stream] Serve received request, start requesting blocktype %v from cID %v to cID %v another client: specific %v block by height %v -> %v",
// 	req.Type,
// 	req.From,
// 	req.To,
// 	req.Specific,
// 	req.Heights[0],
// 	req.Heights[len(req.Heights)-1],
// )
// if err != nil {
// 	logger.Errorf("[stream] No serviceClient with  blocktype %v from cID %v to cID %v another client: specific %v block by height %v -> %v",
// 		req.Type,
// 		req.From,
// 		req.To,
// 		req.Specific,
// 		req.Heights[0],
// 		req.Heights[len(req.Heights)-1],
// 	)
// 	return err
// }
// if err != nil {
// 	logger.Infof("[stream] Calling client but received error %v, return", err)
// 	return err
// }
