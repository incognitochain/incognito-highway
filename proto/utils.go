package proto

import (
	"highway/common"
	"sort"

	"github.com/pkg/errors"
)

func CheckReq(req *BlockByHeightRequest) error {
	if len(req.Heights) < 1 {
		return errors.Errorf("List requested heights is empty")
	}
	if !req.Specific {
		if len(req.Heights) != 2 || req.Heights[1] < req.Heights[0] {
			return errors.Errorf("Invalid requested range blocks, is must be [from,to]")
		}
		if req.Heights[0] == common.GenesisBlockHeight {
			if req.Heights[0] == req.Heights[1] {
				return errors.Errorf("Can not request sync genesis block")
			}
			req.Heights[0]++
		}
	} else {
		sort.Slice(req.Heights, func(i, j int) bool {
			return req.Heights[i] < req.Heights[j]
		})
		if req.Heights[0] == 1 {
			req.Heights[0] = 2
		}
	}
	return nil
}
