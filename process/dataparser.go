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
	// realType := reflect.TypeOf(message)
	// fmt.Printf("Cmd message type of struct %s", realType.String())

	// // cache message hash
	// if peerConn.listenerPeer != nil {
	// 	hashMsg := message.Hash()
	// 	if err := peerConn.listenerPeer.HashToPool(hashMsg); err != nil {
	// 		fmt.Println(err)
	// 		return NewPeerError(CacheMessageHashError, err, nil)
	// 	}
	// }

	// process message for each of message type
	// errProcessMessage := d.processMessageForEachType(realType, message)
	// if errProcessMessage != nil {
	// 	fmt.Println(errProcessMessage)
	// 	return errors.WithStack(errProcessMessage)
	// }
	return message.(*wire.MessagePeerState), nil
}

/*
// NOTE: copy from peerConn.processInMessageString
	// Parse Message header from last 24 bytes header message
	jsonDecodeBytesRaw, err := hex.DecodeString(msgStr)
	if err != nil {
		return errors.Wrapf(err, "msgStr: %v", msgStr)
	}

	// TODO(0xbunyip): separate caching from peerConn
	// // cache message hash
	// hashMsgRaw := common.HashH(jsonDecodeBytesRaw).String()
	// if peerConn.listenerPeer != nil {
	// 	if err := peerConn.listenerPeer.HashToPool(hashMsgRaw); err != nil {
	// 		fmt.Println(err)
	// 		return NewPeerError(HashToPoolError, err, nil)
	// 	}
	// }
	// unzip data before process
	jsonDecodeBytes, err := common.GZipToBytes(jsonDecodeBytesRaw)
	if err != nil {
		fmt.Println("Can not unzip from message")
		fmt.Println(err)
		return errors.WithStack(err)
	}

	fmt.Printf("In message content : %s", string(jsonDecodeBytes))

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
		return errors.WithStack(err)
	}

	if len(jsonDecodeBytes) > message.MaxPayloadLength(wire.Version) {
		fmt.Printf("Msg size exceed MsgType %s max size, size %+v | max allow is %+v \n", commandType, len(jsonDecodeBytes), message.MaxPayloadLength(1))
		return errors.WithStack(err)
	}
	// check forward TODO
	/*if peerConn.config.MessageListeners.GetCurrentRoleShard != nil {
		cRole, cShard := peerConn.config.MessageListeners.GetCurrentRoleShard()
		if cShard != nil {
			fT := messageHeader[wire.MessageCmdTypeSize]
			if fT == MessageToShard {
				fS := messageHeader[wire.MessageCmdTypeSize+1]
				if *cShard != fS {
					if peerConn.config.MessageListeners.PushRawBytesToShard != nil {
						err1 := peerConn.config.MessageListeners.PushRawBytesToShard(peerConn, &jsonDecodeBytesRaw, *cShard)
						if err1 != nil {
							fmt.Println(err1)
						}
					}
					return NewPeerError(CheckForwardError, err, nil)
				}
			}
		}
		if cRole != "" {
			fT := messageHeader[wire.MessageCmdTypeSize]
			if fT == MessageToBeacon && cRole != "beacon" {
				if peerConn.config.MessageListeners.PushRawBytesToBeacon != nil {
					err1 := peerConn.config.MessageListeners.PushRawBytesToBeacon(peerConn, &jsonDecodeBytesRaw)
					if err1 != nil {
						fmt.Println(err1)
					}
				}
				return NewPeerError(CheckForwardError, err, nil)
			}
		}
	}*/
/*
	err = json.Unmarshal(messageBody, &message)
	if err != nil {
		fmt.Println("Can not parse struct from json message")
		fmt.Println(err)
		return errors.WithStack(err)
	}
	realType := reflect.TypeOf(message)
	fmt.Printf("Cmd message type of struct %s", realType.String())

	// // cache message hash
	// if peerConn.listenerPeer != nil {
	// 	hashMsg := message.Hash()
	// 	if err := peerConn.listenerPeer.HashToPool(hashMsg); err != nil {
	// 		fmt.Println(err)
	// 		return NewPeerError(CacheMessageHashError, err, nil)
	// 	}
	// }

	// process message for each of message type
	errProcessMessage := d.processMessageForEachType(realType, message)
	if errProcessMessage != nil {
		fmt.Println(errProcessMessage)
		return errors.WithStack(errProcessMessage)
	}
*/
