package proto

import "highway/common"

func (req *RegisterRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *GetBlockShardByHeightRequest) GetFrom() int32 {
	return req.Shard
}

func (req *GetBlockShardByHeightRequest) GetTo() int32 {
	return req.Shard
}

func (req *GetBlockShardByHeightRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *GetBlockBeaconByHeightRequest) GetFrom() int32 {
	return int32(common.BEACONID)
}

func (req *GetBlockBeaconByHeightRequest) GetTo() int32 {
	return int32(common.BEACONID)
}

func (req *GetBlockBeaconByHeightRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *GetBlockCrossShardByHeightRequest) GetFrom() int32 {
	return req.FromShard
}

func (req *GetBlockCrossShardByHeightRequest) GetTo() int32 {
	return req.ToShard
}

func (req *GetBlockCrossShardByHeightRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *GetBlockShardToBeaconByHeightRequest) GetFrom() int32 {
	return req.FromShard
}

func (req *GetBlockShardToBeaconByHeightRequest) GetTo() int32 {
	return int32(common.BEACONID)
}

func (req *GetBlockShardToBeaconByHeightRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *GetBlockBeaconByHashRequest) GetCID() int32 {
	return int32(common.BEACONID)
}

func (req *GetBlockShardByHashRequest) GetCID() int32 {
	return req.Shard
}

func (req *GetBlockShardByHashRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *GetBlockBeaconByHashRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *GetBlockCrossShardByHashRequest) SetUUID(uuid string) {
	req.UUID = uuid
}

func (req *BlockByHeightRequest) SetUUID(uuid string) {
	req.UUID = uuid
}
