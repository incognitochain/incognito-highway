#!/usr/bin/env bash
protoc -I . register.proto shard_block.proto beacon_block.proto cross_shard.proto --go_out=plugins=grpc:../msg/.
protoc -I . highway.proto --go_out=plugins=grpc,shard_block.proto=msg/,beacon_block.proto=msg/cross_shard.proto=msg/:../process/.
