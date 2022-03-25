package serialize

import (
	"fmt"
	"strconv"

	"go.mau.fi/libsignal/logger"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/util/bytehelper"
	"go.mau.fi/libsignal/util/optional"
	proto "google.golang.org/protobuf/proto"
)

// NewProtoBufSerializer will return a serializer for all Signal objects that will
// be responsible for converting objects to and from ProtoBuf bytes.
func NewProtoBufSerializer() *Serializer {
	serializer := NewSerializer()

	serializer.SignalMessage = &ProtoBufSignalMessageSerializer{}
	serializer.PreKeySignalMessage = &ProtoBufPreKeySignalMessageSerializer{}
	serializer.SenderKeyMessage = &ProtoBufSenderKeyMessageSerializer{}
	serializer.SenderKeyDistributionMessage = &ProtoBufSenderKeyDistributionMessageSerializer{}
	serializer.SignedPreKeyRecord = &JSONSignedPreKeyRecordSerializer{}
	serializer.PreKeyRecord = &JSONPreKeyRecordSerializer{}
	serializer.State = &JSONStateSerializer{}
	serializer.Session = &JSONSessionSerializer{}
	serializer.SenderKeyRecord = &JSONSenderKeySessionSerializer{}
	serializer.SenderKeyState = &JSONSenderKeyStateSerializer{}

	return serializer
}

func highBitsToInt(value byte) int {
	return int((value & 0xFF) >> 4)
}

func intsToByteHighAndLow(highValue, lowValue int) byte {
	return byte((highValue<<4 | lowValue) & 0xFF)
}

// ProtoBufSignalMessageSerializer is a structure for serializing signal messages into
// and from ProtoBuf.
type ProtoBufSignalMessageSerializer struct{}

// Serialize will take a signal message structure and convert it to ProtoBuf bytes.
func (j *ProtoBufSignalMessageSerializer) Serialize(signalMessage *protocol.SignalMessageStructure) []byte {
	sm := &SignalMessage{
		RatchetKey:      signalMessage.RatchetKey,
		Counter:         &signalMessage.Counter,
		PreviousCounter: &signalMessage.PreviousCounter,
		Ciphertext:      signalMessage.CipherText,
	}
	var serialized []byte
	message, err := proto.Marshal(sm)
	if err != nil {
		logger.Error("Error serializing signal message: ", err)
	}

	if signalMessage.Version != 0 {
		serialized = append(serialized, []byte(strconv.Itoa(signalMessage.Version))...)
	}
	serialized = append(serialized, message...)

	if signalMessage.Mac != nil {
		serialized = append(serialized, signalMessage.Mac...)
	}

	return serialized
}

// Deserialize will take in ProtoBuf bytes and return a signal message structure.
func (j *ProtoBufSignalMessageSerializer) Deserialize(serialized []byte) (*protocol.SignalMessageStructure, error) {
	parts, err := bytehelper.SplitThree(serialized, 1, len(serialized)-1-protocol.MacLength, protocol.MacLength)
	if err != nil {
		logger.Error("Error split signal message: ", err)
		return nil, err
	}
	version := highBitsToInt(parts[0][0])
	message := parts[1]
	mac := parts[2]

	var sm SignalMessage
	err = proto.Unmarshal(message, &sm)
	if err != nil {
		logger.Error("Error deserializing signal message: ", err)
		return nil, err
	}

	signalMessage := protocol.SignalMessageStructure{
		Version:         version,
		RatchetKey:      sm.GetRatchetKey(),
		Counter:         sm.GetCounter(),
		PreviousCounter: sm.GetPreviousCounter(),
		CipherText:      sm.GetCiphertext(),
		Mac:             mac,
	}

	return &signalMessage, nil
}

// ProtoBufPreKeySignalMessageSerializer is a structure for serializing prekey signal messages
// into and from ProtoBuf.
type ProtoBufPreKeySignalMessageSerializer struct{}

// Serialize will take a prekey signal message structure and convert it to ProtoBuf bytes.
func (j *ProtoBufPreKeySignalMessageSerializer) Serialize(signalMessage *protocol.PreKeySignalMessageStructure) []byte {
	preKeyMessage := &PreKeySignalMessage{
		RegistrationId: &signalMessage.RegistrationID,
		SignedPreKeyId: &signalMessage.SignedPreKeyID,
		BaseKey:        signalMessage.BaseKey,
		IdentityKey:    signalMessage.IdentityKey,
		Message:        signalMessage.Message,
	}

	if !signalMessage.PreKeyID.IsEmpty {
		preKeyMessage.PreKeyId = &signalMessage.PreKeyID.Value
	}

	message, err := proto.Marshal(preKeyMessage)
	if err != nil {
		logger.Error("Error serializing prekey signal message: ", err)
	}

	serialized := append([]byte(strconv.Itoa(signalMessage.Version)), message...)
	logger.Debug("Serialize PreKeySignalMessage result: ", serialized)
	return serialized
}

// Deserialize will take in ProtoBuf bytes and return a prekey signal message structure.
func (j *ProtoBufPreKeySignalMessageSerializer) Deserialize(serialized []byte) (*protocol.PreKeySignalMessageStructure, error) {
	version := highBitsToInt(serialized[0])
	message := serialized[1:]
	var sm PreKeySignalMessage
	err := proto.Unmarshal(message, &sm)
	if err != nil {
		logger.Error("Error deserializing prekey signal message: ", err)
		return nil, err
	}

	preKeyId := optional.NewEmptyUint32()
	if sm.GetPreKeyId() != 0 {
		preKeyId = optional.NewOptionalUint32(sm.GetPreKeyId())
	}

	preKeySignalMessage := protocol.PreKeySignalMessageStructure{
		Version:        version,
		RegistrationID: sm.GetRegistrationId(),
		BaseKey:        sm.GetBaseKey(),
		IdentityKey:    sm.GetIdentityKey(),
		SignedPreKeyID: sm.GetSignedPreKeyId(),
		Message:        sm.GetMessage(),
		PreKeyID:       preKeyId,
	}

	return &preKeySignalMessage, nil
}

// ProtoBufSenderKeyDistributionMessageSerializer is a structure for serializing senderkey
// distribution records to and from ProtoBuf.
type ProtoBufSenderKeyDistributionMessageSerializer struct{}

// Serialize will take a senderkey distribution message and convert it to ProtoBuf bytes.
func (j *ProtoBufSenderKeyDistributionMessageSerializer) Serialize(message *protocol.SenderKeyDistributionMessageStructure) []byte {
	senderDis := SenderKeyDistributionMessage{
		Id:         &message.ID,
		Iteration:  &message.Iteration,
		ChainKey:   message.ChainKey,
		SigningKey: message.SigningKey,
	}

	serialized, err := proto.Marshal(&senderDis)
	if err != nil {
		logger.Error("Error serializing senderkey distribution message: ", err)
	}

	version := strconv.Itoa(int(message.Version))
	serialized = append([]byte(version), serialized...)
	logger.Debug("Serialize result: ", serialized)
	return serialized
}

// Deserialize will take in ProtoBuf bytes and return a message structure, which can be
// used to create a new SenderKey Distribution object.
func (j *ProtoBufSenderKeyDistributionMessageSerializer) Deserialize(serialized []byte) (*protocol.SenderKeyDistributionMessageStructure, error) {
	version := uint32(highBitsToInt(serialized[0]))
	message := serialized[1:]

	var senderKeyDis SenderKeyDistributionMessage
	err := proto.Unmarshal(message, &senderKeyDis)
	if err != nil {
		logger.Error("Error deserializing senderkey distribution message: ", err)
		return nil, err
	}

	msgStructure := protocol.SenderKeyDistributionMessageStructure{
		ID:         senderKeyDis.GetId(),
		Iteration:  senderKeyDis.GetIteration(),
		ChainKey:   senderKeyDis.GetChainKey(),
		SigningKey: senderKeyDis.GetSigningKey(),
		Version:    version,
	}
	return &msgStructure, nil
}

// ProtoBufSenderKeyMessageSerializer is a structure for serializing senderkey
// messages to and from ProtoBuf.
type ProtoBufSenderKeyMessageSerializer struct{}

// Serialize will take a senderkey message and convert it to ProtoBuf bytes.
func (j *ProtoBufSenderKeyMessageSerializer) Serialize(message *protocol.SenderKeyMessageStructure) []byte {
	senderMessage := &SenderKeyMessage{
		Id:         &message.ID,
		Iteration:  &message.Iteration,
		Ciphertext: message.CipherText,
	}

	var serialized []byte
	m, err := proto.Marshal(senderMessage)
	if err != nil {
		logger.Error("Error serializing signal message: ", err)
	}

	if message.Version != 0 {
		serialized = append([]byte(fmt.Sprint(message.Version)), m...)
	}

	if message.Signature != nil {
		serialized = append(serialized, message.Signature...)
	}
	logger.Debug("Serialize result: ", serialized)
	return serialized
}

// Deserialize will take in ProtoBuf bytes and return a message structure, which can be
// used to create a new SenderKey message object.
func (j *ProtoBufSenderKeyMessageSerializer) Deserialize(serialized []byte) (*protocol.SenderKeyMessageStructure, error) {
	parts, err := bytehelper.SplitThree(serialized, 1, len(serialized)-1-64, 64)
	if err != nil {
		logger.Error("Error split signal message: ", err)
		return nil, err
	}
	version := uint32(highBitsToInt(parts[0][0]))
	message := parts[1]
	signature := parts[2]

	var senderKey SenderKeyMessage
	err = proto.Unmarshal(message, &senderKey)
	if err != nil {
		logger.Error("Error deserializing senderkey message: ", err)
		return nil, err
	}

	msgStructure := protocol.SenderKeyMessageStructure{
		Version:    version,
		ID:         senderKey.GetId(),
		Iteration:  senderKey.GetIteration(),
		CipherText: senderKey.GetCiphertext(),
		Signature:  signature,
	}

	return &msgStructure, nil
}
