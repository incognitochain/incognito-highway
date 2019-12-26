package p2p

import "time"

const (
	GRPCMaxConnectionIdle = 1 * time.Hour
	GRPCTime              = 10 * time.Minute
	GRPCTimeout           = 10 * time.Second
)
