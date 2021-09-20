package common

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

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

func ParseMsgBody(data []byte) ([]byte, string, error) {
	dataStr := string(data)
	jsonDecodeBytesRaw, err := hex.DecodeString(dataStr)
	if err != nil {
		return nil, "", errors.Wrapf(err, "msgStr: %v", dataStr)
	}
	jsonDecodeBytes, err := GZipToBytes(jsonDecodeBytesRaw)
	if err != nil {
		fmt.Println("Can not unzip from message")
		fmt.Println(err)
		return nil, "", errors.WithStack(err)
	}
	// Parse Message body
	messageBody := jsonDecodeBytes[:len(jsonDecodeBytes)-wire.MessageHeaderSize]
	messageHeader := jsonDecodeBytes[len(jsonDecodeBytes)-wire.MessageHeaderSize:]

	// get cmd type in header message
	commandInHeader := bytes.Trim(messageHeader[:wire.MessageCmdTypeSize], "\x00")
	commandType := string(messageHeader[:len(commandInHeader)])
	// convert to particular message from message cmd type

	return messageBody, commandType, nil
}

func GetTransactionTimestamp(data []byte) (int64, error) {
	messageBody, msgType, err := ParseMsgBody(data)
	if err != nil {
		return 0, err
	}
	if (msgType == wire.CmdTx) || (msgType == wire.CmdPrivacyCustomToken) {
		message := &struct {
			Transaction json.RawMessage
		}{}
		err = json.Unmarshal(messageBody, message)
		if (err != nil) || (message.Transaction == nil) {
			fmt.Println("Can not parse struct from json message")
			fmt.Println(err)
			return 0, errors.WithStack(err)
		}
		rawTx := make(map[string]interface{})
		err = json.Unmarshal(message.Transaction, &rawTx)
		if err != nil {
			return 0, err
		}
		if locktimeI, hasLocktime := rawTx["LockTime"]; hasLocktime {
			if lockTime, ok := locktimeI.(float64); ok {
				return int64(lockTime), nil
			}
			return 0, errors.Errorf("Can cast %v to float64", locktimeI)
		}
		return 0, errors.Errorf("Can not get lock time from raw tx %v", message.Transaction)
	}

	return 0, errors.Errorf("Can not get tx timestamp from msg type %v ", msgType)
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

func FilterLocalAddrs(mas []multiaddr.Multiaddr) []multiaddr.Multiaddr {
	localAddrs := []string{
		"127.0.0.1",
		"0.0.0.0",
		"192.168.",
		"/ip4/172.",
	}
	nonLocal := []multiaddr.Multiaddr{}
	for _, ma := range mas {
		local := false
		for _, s := range localAddrs {
			if strings.Contains(ma.String(), s) {
				local = true
				break
			}
		}
		if !local {
			nonLocal = append(nonLocal, ma)
		}
	}
	return nonLocal
}
