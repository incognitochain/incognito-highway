package proto

import "sort"

func CheckReq(req *BlockByHeightRequest) bool {
	if len(req.Heights) < 1 {
		return false
	}
	if !req.Specific {
		if len(req.Heights) != 2 || req.Heights[1] < req.Heights[0] {
			return false
		}
		if req.Heights[0] == 1 {
			if req.Heights[0] == req.Heights[1] {
				return false
			}
			req.Heights[0] = 2
		}
	} else {
		sort.Slice(req.Heights, func(i, j int) bool {
			return req.Heights[i] < req.Heights[j]
		})
		if req.Heights[0] == 1 {
			req.Heights[0] = 2
		}
	}
	return true
}
