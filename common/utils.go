package common

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
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

func ParsePeerStateData(dataStr string) (*wire.MessagePeerState, error) {
	jsonDecodeBytesRaw, err := hex.DecodeString(dataStr)
	if err != nil {
		return nil, errors.Wrapf(err, "msgStr: %v", dataStr)
	}
	jsonDecodeBytes, err := GZipToBytes(jsonDecodeBytesRaw)
	if err != nil {
		fmt.Println("Can not unzip from message")
		fmt.Println(err)
		return nil, errors.WithStack(err)
	}
	// Parse Message body
	messageBody := jsonDecodeBytes[:len(jsonDecodeBytes)-wire.MessageHeaderSize]
	messageHeader := jsonDecodeBytes[len(jsonDecodeBytes)-wire.MessageHeaderSize:]

	// get cmd type in header message
	commandInHeader := bytes.Trim(messageHeader[:wire.MessageCmdTypeSize], "\x00")
	commandType := string(messageHeader[:len(commandInHeader)])
	// convert to particular message from message cmd type
	message, err := wire.MakeEmptyMessage(string(commandType))
	if err != nil {
		fmt.Println("Can not find particular message for message cmd type")
		fmt.Println(err)
		return nil, errors.WithStack(err)
	}

	if len(jsonDecodeBytes) > message.MaxPayloadLength(wire.Version) {
		fmt.Printf("Msg size exceed MsgType %s max size, size %+v | max allow is %+v \n", commandType, len(jsonDecodeBytes), message.MaxPayloadLength(1))
		return nil, errors.WithStack(err)
	}
	err = json.Unmarshal(messageBody, &message)
	if err != nil {
		fmt.Println("Can not parse struct from json message")
		fmt.Println(err)
		return nil, errors.WithStack(err)
	}
	return message.(*wire.MessagePeerState), nil
}

func GZipToBytes(src []byte) ([]byte, error) {
	reader := bytes.NewReader(src)
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	resultBytes, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	return resultBytes, nil
}

func NewKeyForCacheDataOfTopic(topic string, data []byte) []byte {
	res := append([]byte(topic), data...)
	return common.HashB(res)
}

func DeleteStringInList(key string, list []string) []string {
	j := 0
	for _, v := range list {
		if v == key {
			continue
		}
		list[j] = v
		j++
	}
	return list[:j]
}

func ParseMsgChainData(data []byte) (wire.Message, error) {
	dataStr := string(data)
	jsonDecodeBytesRaw, err := hex.DecodeString(dataStr)
	if err != nil {
		return nil, errors.Wrapf(err, "msgStr: %v", dataStr)
	}
	jsonDecodeBytes, err := GZipToBytes(jsonDecodeBytesRaw)
	if err != nil {
		fmt.Println("Can not unzip from message")
		fmt.Println(err)
		return nil, errors.WithStack(err)
	}
	// Parse Message body
	messageBody := jsonDecodeBytes[:len(jsonDecodeBytes)-wire.MessageHeaderSize]
	messageHeader := jsonDecodeBytes[len(jsonDecodeBytes)-wire.MessageHeaderSize:]

	// get cmd type in header message
	commandInHeader := bytes.Trim(messageHeader[:wire.MessageCmdTypeSize], "\x00")
	commandType := string(messageHeader[:len(commandInHeader)])
	// convert to particular message from message cmd type
	message, err := wire.MakeEmptyMessage(string(commandType))
	if err != nil {
		fmt.Println("Can not find particular message for message cmd type")
		fmt.Println(err)
		return nil, errors.WithStack(err)
	}

	if len(jsonDecodeBytes) > message.MaxPayloadLength(wire.Version) {
		fmt.Printf("Msg size exceed MsgType %s max size, size %+v | max allow is %+v \n", commandType, len(jsonDecodeBytes), message.MaxPayloadLength(1))
		return nil, errors.WithStack(err)
	}
	err = json.Unmarshal(messageBody, &message)
	if err != nil {
		fmt.Println("Can not parse struct from json message")
		fmt.Println(err)
		return nil, errors.WithStack(err)
	}
	return message, nil
}
