package bsoulseek

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
)

type soulseekMessage interface {}
type soulseekMessageResponse interface {}

const (
	loginMessageCode          uint32 = 1
	sayInChatRoomMessageCode  uint32 = 13
	joinRoomMessageCode       uint32 = 14
	userJoinedRoomMessageCode uint32 = 16
	userLeftRoomMessageCode   uint32 = 17
	privateMessageCode        uint32 = 22
	kickedMessageCode         uint32 = 41
	changePasswordMessageCode  uint32 = 142
)

var ignoreMessageCodes = map[uint32]bool {
	7:   true,
	64:  true,
	69:  true,
	83:  true,
	84:  true,
	104: true,
	113: true,
	114: true,
	115: true,
	130: true,
	133: true,
	139: true,
	140: true,
	145: true,
	146: true,
	148: true,
	160: true,
	1003: true,
}


// 1: Login
type loginMessage struct {
	Code         uint32
	Username     string
	Password     string
	Version      uint32
	Hash         string
	MinorVersion uint32
}

type loginMessageResponseSuccess struct {
	Greet       string
	Address     uint32
	Hash        string
	IsSupporter bool
}

type loginMessageResponseFailure struct {
	Reason string
}


// 13: Say in chatroom
type sayChatroomMessage struct {
	Code    uint32
	Room string
	Message string
}

type sayChatroomMessageReceive struct {
	Room string
	Username string
	Message string
}


// 14: Join room
type joinRoomMessage struct {
	Code    uint32
	Room    string
	Private uint32
}

type userStat struct {
	AvgSpeed  uint32
	UploadNum uint64
	Files     uint32
	Dirs      uint32
}

type joinRoomMessageResponse struct {
	Room      string
	Users     []string
	Statuses  []uint32
	Stats     []userStat
	SlotsFree []uint32
	Countries []uint32
	Owner     string
	Operators []string
}


// 16: User joined room
type userJoinedRoomMessage struct {
	Room string
	Username string
	Status uint32
	AvgSpeed uint32
	UploadNum uint64
	Files uint32
	Dirs uint32
	SlotsFree uint32
	CountryCode string
}


// 16: User left room
type userLeftRoomMessage struct {
	Room string
	Username string
}


// 22: Private message (sometimes used by server to tell us errors)
type privateMessageReceive struct {
	ID uint32
	Timestamp uint32
	Username string
	Message string
	NewMessage bool
}


// 41: Kicked from server (relogged)
type kickedMessageResponse struct {}


// 142: Change password
type changePasswordMessage struct {
	Code uint32
	Password string
}

type changePasswordMessageResponse struct {
	Password string
}


func packMessage(message soulseekMessage) ([]byte, error) {
	buf := &bytes.Buffer{}
	var length uint32 = 0
	binary.Write(buf, binary.LittleEndian, length) // placeholder
	v := reflect.ValueOf(message)
	var err error
	for i := range v.NumField() {
		val := v.Field(i).Interface()
		switch val := val.(type) {
		case string:
			s_len := uint32(len(val))
			err = binary.Write(buf, binary.LittleEndian, s_len)
			buf.WriteString(val)
			length += s_len + 4
		case bool, uint8:
			length += 1
			err = binary.Write(buf, binary.LittleEndian, val)
		case uint16:
			length += 2
			err = binary.Write(buf, binary.LittleEndian, val)
		case uint32:
			length += 4
			err = binary.Write(buf, binary.LittleEndian, val)
		case uint64:
			length += 8
			err = binary.Write(buf, binary.LittleEndian, val)
		default:
			panic("Unsupported struct field type")
		}
		if err != nil {
			return nil, err
		}
	}
	bytes := buf.Bytes()
	binary.LittleEndian.PutUint32(bytes, length)
	return bytes, nil
}

func unpackStructField(reader io.Reader, field reflect.Value) error {
	switch field.Kind() {
	case reflect.Struct:
		for i := range field.NumField() {
			field := field.Field(i)
			err := unpackStructField(reader, field)
			if err != nil {
				return err
			}
		}
	case reflect.Slice:
		var length uint32
		err := binary.Read(reader, binary.LittleEndian, &length)
		if err != nil {
			return err
		}
		ilen := int(length)
		newval := reflect.MakeSlice(field.Type(), ilen, ilen)
		field.Set(newval)
		for j := range ilen {
			err := unpackStructField(reader, field.Index(j))
			if err != nil {
				return err
			}
		}
	case reflect.String:
		var length uint32
		err := binary.Read(reader, binary.LittleEndian, &length)
		if err != nil {
			return err
		}
		val := make([]byte, length)
		_, err = reader.Read(val)
		if err != nil {
			return err
		}
		field.SetString(string(val))
	case reflect.Bool:
		var val bool
		err := binary.Read(reader, binary.LittleEndian, &val)
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Uint8:
		var val uint8
		err := binary.Read(reader, binary.LittleEndian, &val)
		if err != nil {
			return err
		}
		field.SetUint(uint64(val))
	case reflect.Uint16:
		var val uint16
		err := binary.Read(reader, binary.LittleEndian, &val)
		if err != nil {
			return err
		}
		field.SetUint(uint64(val))
	case reflect.Uint32:
		var val uint32
		err := binary.Read(reader, binary.LittleEndian, &val)
		if err != nil {
			return err
		}
		field.SetUint(uint64(val))
	case reflect.Uint64:
		var val uint64
		err := binary.Read(reader, binary.LittleEndian, &val)
		if err != nil {
			return err
		}
		field.SetUint(val)
	default:
		panic(fmt.Sprintf("Unsupported struct field type: %d", field.Kind()))
	}
	return nil
}

func unpackMessage[T soulseekMessage](reader io.Reader) (T, error) {
	var data T
	v := reflect.ValueOf(&data).Elem()
	err := unpackStructField(reader, v)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return data, err
	}
	return data, nil
}

func readMessage(reader io.Reader) (soulseekMessage, error) {
	var length uint32
	err := binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(length))
	_, err = io.ReadAtLeast(reader, buf, len(buf))
	if err != nil {
		return nil, err
	}

	reader = bytes.NewReader(buf)

	var code uint32
	err = binary.Read(reader, binary.LittleEndian, &code)
	if err != nil {
		return nil, err
	}
	switch code {
	case loginMessageCode:
		// login message is special; has two possible responses
		var success bool
		err := binary.Read(reader, binary.LittleEndian, &success)
		if err != nil {
			return nil, err
		}
		if success {
			return unpackMessage[loginMessageResponseSuccess](reader)
		} else {
			return unpackMessage[loginMessageResponseFailure](reader)
		}
	case sayInChatRoomMessageCode:
		return unpackMessage[sayChatroomMessageReceive](reader)
	case joinRoomMessageCode:
		return unpackMessage[joinRoomMessageResponse](reader)
	case kickedMessageCode:
		return unpackMessage[kickedMessageResponse](reader)
	case userJoinedRoomMessageCode:
		return unpackMessage[userJoinedRoomMessage](reader)
	case userLeftRoomMessageCode:
		return unpackMessage[userLeftRoomMessage](reader)
	case changePasswordMessageCode:
		return unpackMessage[changePasswordMessageResponse](reader)
	case privateMessageCode:
		return unpackMessage[privateMessageReceive](reader)
	default:
		_, ignore := ignoreMessageCodes[code]
		if ignore {
			return nil, nil
		}
		return nil, fmt.Errorf("Unknown message code: %d", code)
	}
}

func makeLoginMessage(username string, password string) soulseekMessage {
	hash := md5.Sum([]byte(username + password))
	msg := loginMessage{
		loginMessageCode,
		username,
		password,
		160,
		hex.EncodeToString(hash[:]),
		1,
	}
	return msg
}

func makeJoinRoomMessage(room string) joinRoomMessage {
	return joinRoomMessage{
		joinRoomMessageCode,
		room,
		0,
	}
}

func makeSayChatroomMessage(room string, text string) sayChatroomMessage {
	return sayChatroomMessage{
		sayInChatRoomMessageCode,
		room,
		text,
	}
}