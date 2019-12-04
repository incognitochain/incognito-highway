package common

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

type Key struct {
	Payment         string `json:"PaymentAddress"`
	CommitteePubKey string `json:"CommitteePublicKey"`
}

type KeyList struct {
	Bc     []Key         `json:"Beacon"`
	Sh     map[int][]Key `json:"Shard"`
	ShPend map[int][]Key `json:"ShardPending"`
}

func HasValuesAt(
	slice []byte,
	value byte,
) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func HasStringAt(
	slice []string,
	value string,
) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func StringToAddrInfo(ma string) (*peer.AddrInfo, error) {
	addr, err := multiaddr.NewMultiaddr(ma)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return addrInfo, nil
}

func NewDefaultMarshaler(data interface{}) json.Marshaler {
	return &marshaler{data}
}

type marshaler struct {
	data interface{}
}

var _ json.Marshaler = (*marshaler)(nil)

func (m *marshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}

func BytesToInts(b []byte) []int {
	s := make([]int, len(b))
	for i, v := range b {
		s[i] = int(v)
	}
	return s
}
