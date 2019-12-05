package rpcserver

type Response struct {
	PeerPerShard map[string][]string
}

type Request struct {
	Shard []string
}
