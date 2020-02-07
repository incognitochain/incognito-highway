package proto

import "highway/common"

func (req *GetBlockShardByHeightRequest) GetFrom() int32 {
	return req.Shard
}

func (req *GetBlockShardByHeightRequest) GetTo() int32 {
	return req.Shard
}

func (req *GetBlockBeaconByHeightRequest) GetFrom() int32 {
	return int32(common.BEACONID)
}

func (req *GetBlockBeaconByHeightRequest) GetTo() int32 {
	return int32(common.BEACONID)
}

func (req *GetBlockCrossShardByHeightRequest) GetFrom() int32 {
	return req.FromShard
}

func (req *GetBlockCrossShardByHeightRequest) GetTo() int32 {
	return req.ToShard
}

func (req *GetBlockShardToBeaconByHeightRequest) GetFrom() int32 {
	return req.FromShard
}

func (req *GetBlockShardToBeaconByHeightRequest) GetTo() int32 {
	return int32(common.BEACONID)
}
