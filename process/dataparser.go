package process

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	fmt "fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/pkg/errors"
)

func ParsePeerStateData(dataStr string) (*wire.MessagePeerState, error) {
	jsonDecodeBytesRaw, err := hex.DecodeString(dataStr)
	if err != nil {
		return nil, errors.Wrapf(err, "msgStr: %v", dataStr)
	}
	jsonDecodeBytes, err := common.GZipToBytes(jsonDecodeBytesRaw)
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
