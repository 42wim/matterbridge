// Generated code
// DO NOT EDIT

package steamlang

import (
	"encoding/binary"
	. "github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/Philipp15b/go-steam/rwu"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/golang/protobuf/proto"
	"io"
)

const UdpHeader_MAGIC uint32 = 0x31305356

type UdpHeader struct {
	Magic        uint32
	PayloadSize  uint16
	PacketType   EUdpPacketType
	Flags        uint8
	SourceConnID uint32
	DestConnID   uint32
	SeqThis      uint32
	SeqAck       uint32
	PacketsInMsg uint32
	MsgStartSeq  uint32
	MsgSize      uint32
}

func NewUdpHeader() *UdpHeader {
	return &UdpHeader{
		Magic:        UdpHeader_MAGIC,
		PacketType:   EUdpPacketType_Invalid,
		SourceConnID: 512,
	}
}

func (d *UdpHeader) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Magic)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.PayloadSize)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.PacketType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Flags)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SourceConnID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.DestConnID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SeqThis)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SeqAck)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.PacketsInMsg)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.MsgStartSeq)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.MsgSize)
	return err
}

func (d *UdpHeader) Deserialize(r io.Reader) error {
	var err error
	d.Magic, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.PayloadSize, err = rwu.ReadUint16(r)
	if err != nil {
		return err
	}
	t0, err := rwu.ReadUint8(r)
	if err != nil {
		return err
	}
	d.PacketType = EUdpPacketType(t0)
	d.Flags, err = rwu.ReadUint8(r)
	if err != nil {
		return err
	}
	d.SourceConnID, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.DestConnID, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.SeqThis, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.SeqAck, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.PacketsInMsg, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.MsgStartSeq, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.MsgSize, err = rwu.ReadUint32(r)
	return err
}

const ChallengeData_CHALLENGE_MASK uint32 = 0xA426DF2B

type ChallengeData struct {
	ChallengeValue uint32
	ServerLoad     uint32
}

func NewChallengeData() *ChallengeData {
	return &ChallengeData{}
}

func (d *ChallengeData) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.ChallengeValue)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ServerLoad)
	return err
}

func (d *ChallengeData) Deserialize(r io.Reader) error {
	var err error
	d.ChallengeValue, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.ServerLoad, err = rwu.ReadUint32(r)
	return err
}

const ConnectData_CHALLENGE_MASK uint32 = ChallengeData_CHALLENGE_MASK

type ConnectData struct {
	ChallengeValue uint32
}

func NewConnectData() *ConnectData {
	return &ConnectData{}
}

func (d *ConnectData) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.ChallengeValue)
	return err
}

func (d *ConnectData) Deserialize(r io.Reader) error {
	var err error
	d.ChallengeValue, err = rwu.ReadUint32(r)
	return err
}

type Accept struct {
}

func NewAccept() *Accept {
	return &Accept{}
}

func (d *Accept) Serialize(w io.Writer) error {
	var err error
	return err
}

func (d *Accept) Deserialize(r io.Reader) error {
	var err error
	return err
}

type Datagram struct {
}

func NewDatagram() *Datagram {
	return &Datagram{}
}

func (d *Datagram) Serialize(w io.Writer) error {
	var err error
	return err
}

func (d *Datagram) Deserialize(r io.Reader) error {
	var err error
	return err
}

type Disconnect struct {
}

func NewDisconnect() *Disconnect {
	return &Disconnect{}
}

func (d *Disconnect) Serialize(w io.Writer) error {
	var err error
	return err
}

func (d *Disconnect) Deserialize(r io.Reader) error {
	var err error
	return err
}

type MsgHdr struct {
	Msg         EMsg
	TargetJobID uint64
	SourceJobID uint64
}

func NewMsgHdr() *MsgHdr {
	return &MsgHdr{
		Msg:         EMsg_Invalid,
		TargetJobID: ^uint64(0),
		SourceJobID: ^uint64(0),
	}
}

func (d *MsgHdr) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Msg)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.TargetJobID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SourceJobID)
	return err
}

func (d *MsgHdr) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Msg = EMsg(t0)
	d.TargetJobID, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SourceJobID, err = rwu.ReadUint64(r)
	return err
}

type ExtendedClientMsgHdr struct {
	Msg           EMsg
	HeaderSize    uint8
	HeaderVersion uint16
	TargetJobID   uint64
	SourceJobID   uint64
	HeaderCanary  uint8
	SteamID       steamid.SteamId
	SessionID     int32
}

func NewExtendedClientMsgHdr() *ExtendedClientMsgHdr {
	return &ExtendedClientMsgHdr{
		Msg:           EMsg_Invalid,
		HeaderSize:    36,
		HeaderVersion: 2,
		TargetJobID:   ^uint64(0),
		SourceJobID:   ^uint64(0),
		HeaderCanary:  239,
	}
}

func (d *ExtendedClientMsgHdr) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Msg)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.HeaderSize)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.HeaderVersion)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.TargetJobID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SourceJobID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.HeaderCanary)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SessionID)
	return err
}

func (d *ExtendedClientMsgHdr) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Msg = EMsg(t0)
	d.HeaderSize, err = rwu.ReadUint8(r)
	if err != nil {
		return err
	}
	d.HeaderVersion, err = rwu.ReadUint16(r)
	if err != nil {
		return err
	}
	d.TargetJobID, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SourceJobID, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.HeaderCanary, err = rwu.ReadUint8(r)
	if err != nil {
		return err
	}
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamID = steamid.SteamId(t1)
	d.SessionID, err = rwu.ReadInt32(r)
	return err
}

type MsgHdrProtoBuf struct {
	Msg          EMsg
	HeaderLength int32
	Proto        *CMsgProtoBufHeader
}

func NewMsgHdrProtoBuf() *MsgHdrProtoBuf {
	return &MsgHdrProtoBuf{
		Msg:   EMsg_Invalid,
		Proto: new(CMsgProtoBufHeader),
	}
}

func (d *MsgHdrProtoBuf) Serialize(w io.Writer) error {
	var err error
	buf0, err := proto.Marshal(d.Proto)
	if err != nil {
		return err
	}
	d.HeaderLength = int32(len(buf0))
	err = binary.Write(w, binary.LittleEndian, EMsg(uint32(d.Msg)|ProtoMask))
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.HeaderLength)
	if err != nil {
		return err
	}
	_, err = w.Write(buf0)
	return err
}

func (d *MsgHdrProtoBuf) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Msg = EMsg(uint32(t0) & EMsgMask)
	d.HeaderLength, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	buf1 := make([]byte, d.HeaderLength, d.HeaderLength)
	_, err = io.ReadFull(r, buf1)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(buf1, d.Proto)
	return err
}

type MsgGCHdrProtoBuf struct {
	Msg          uint32
	HeaderLength int32
	Proto        *CMsgProtoBufHeader
}

func NewMsgGCHdrProtoBuf() *MsgGCHdrProtoBuf {
	return &MsgGCHdrProtoBuf{
		Msg:   0,
		Proto: new(CMsgProtoBufHeader),
	}
}

func (d *MsgGCHdrProtoBuf) Serialize(w io.Writer) error {
	var err error
	buf0, err := proto.Marshal(d.Proto)
	if err != nil {
		return err
	}
	d.HeaderLength = int32(len(buf0))
	err = binary.Write(w, binary.LittleEndian, EMsg(uint32(d.Msg)|ProtoMask))
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.HeaderLength)
	if err != nil {
		return err
	}
	_, err = w.Write(buf0)
	return err
}

func (d *MsgGCHdrProtoBuf) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.Msg = uint32(t0) & EMsgMask
	d.HeaderLength, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	buf1 := make([]byte, d.HeaderLength, d.HeaderLength)
	_, err = io.ReadFull(r, buf1)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(buf1, d.Proto)
	return err
}

type MsgGCHdr struct {
	HeaderVersion uint16
	TargetJobID   uint64
	SourceJobID   uint64
}

func NewMsgGCHdr() *MsgGCHdr {
	return &MsgGCHdr{
		HeaderVersion: 1,
		TargetJobID:   ^uint64(0),
		SourceJobID:   ^uint64(0),
	}
}

func (d *MsgGCHdr) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.HeaderVersion)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.TargetJobID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SourceJobID)
	return err
}

func (d *MsgGCHdr) Deserialize(r io.Reader) error {
	var err error
	d.HeaderVersion, err = rwu.ReadUint16(r)
	if err != nil {
		return err
	}
	d.TargetJobID, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SourceJobID, err = rwu.ReadUint64(r)
	return err
}

type MsgClientJustStrings struct {
}

func NewMsgClientJustStrings() *MsgClientJustStrings {
	return &MsgClientJustStrings{}
}

func (d *MsgClientJustStrings) GetEMsg() EMsg {
	return EMsg_Invalid
}

func (d *MsgClientJustStrings) Serialize(w io.Writer) error {
	var err error
	return err
}

func (d *MsgClientJustStrings) Deserialize(r io.Reader) error {
	var err error
	return err
}

type MsgClientGenericResponse struct {
	Result EResult
}

func NewMsgClientGenericResponse() *MsgClientGenericResponse {
	return &MsgClientGenericResponse{}
}

func (d *MsgClientGenericResponse) GetEMsg() EMsg {
	return EMsg_Invalid
}

func (d *MsgClientGenericResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	return err
}

func (d *MsgClientGenericResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	d.Result = EResult(t0)
	return err
}

const MsgChannelEncryptRequest_PROTOCOL_VERSION uint32 = 1

type MsgChannelEncryptRequest struct {
	ProtocolVersion uint32
	Universe        EUniverse
}

func NewMsgChannelEncryptRequest() *MsgChannelEncryptRequest {
	return &MsgChannelEncryptRequest{
		ProtocolVersion: MsgChannelEncryptRequest_PROTOCOL_VERSION,
		Universe:        EUniverse_Invalid,
	}
}

func (d *MsgChannelEncryptRequest) GetEMsg() EMsg {
	return EMsg_ChannelEncryptRequest
}

func (d *MsgChannelEncryptRequest) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.ProtocolVersion)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Universe)
	return err
}

func (d *MsgChannelEncryptRequest) Deserialize(r io.Reader) error {
	var err error
	d.ProtocolVersion, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	t0, err := rwu.ReadInt32(r)
	d.Universe = EUniverse(t0)
	return err
}

type MsgChannelEncryptResponse struct {
	ProtocolVersion uint32
	KeySize         uint32
}

func NewMsgChannelEncryptResponse() *MsgChannelEncryptResponse {
	return &MsgChannelEncryptResponse{
		ProtocolVersion: MsgChannelEncryptRequest_PROTOCOL_VERSION,
		KeySize:         128,
	}
}

func (d *MsgChannelEncryptResponse) GetEMsg() EMsg {
	return EMsg_ChannelEncryptResponse
}

func (d *MsgChannelEncryptResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.ProtocolVersion)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.KeySize)
	return err
}

func (d *MsgChannelEncryptResponse) Deserialize(r io.Reader) error {
	var err error
	d.ProtocolVersion, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.KeySize, err = rwu.ReadUint32(r)
	return err
}

type MsgChannelEncryptResult struct {
	Result EResult
}

func NewMsgChannelEncryptResult() *MsgChannelEncryptResult {
	return &MsgChannelEncryptResult{
		Result: EResult_Invalid,
	}
}

func (d *MsgChannelEncryptResult) GetEMsg() EMsg {
	return EMsg_ChannelEncryptResult
}

func (d *MsgChannelEncryptResult) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	return err
}

func (d *MsgChannelEncryptResult) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	d.Result = EResult(t0)
	return err
}

type MsgClientNewLoginKey struct {
	UniqueID uint32
	LoginKey []uint8
}

func NewMsgClientNewLoginKey() *MsgClientNewLoginKey {
	return &MsgClientNewLoginKey{
		LoginKey: make([]uint8, 20, 20),
	}
}

func (d *MsgClientNewLoginKey) GetEMsg() EMsg {
	return EMsg_ClientNewLoginKey
}

func (d *MsgClientNewLoginKey) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.UniqueID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.LoginKey)
	return err
}

func (d *MsgClientNewLoginKey) Deserialize(r io.Reader) error {
	var err error
	d.UniqueID, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, d.LoginKey)
	return err
}

type MsgClientNewLoginKeyAccepted struct {
	UniqueID uint32
}

func NewMsgClientNewLoginKeyAccepted() *MsgClientNewLoginKeyAccepted {
	return &MsgClientNewLoginKeyAccepted{}
}

func (d *MsgClientNewLoginKeyAccepted) GetEMsg() EMsg {
	return EMsg_ClientNewLoginKeyAccepted
}

func (d *MsgClientNewLoginKeyAccepted) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.UniqueID)
	return err
}

func (d *MsgClientNewLoginKeyAccepted) Deserialize(r io.Reader) error {
	var err error
	d.UniqueID, err = rwu.ReadUint32(r)
	return err
}

const (
	MsgClientLogon_ObfuscationMask                                      uint32 = 0xBAADF00D
	MsgClientLogon_CurrentProtocol                                      uint32 = 65580
	MsgClientLogon_ProtocolVerMajorMask                                 uint32 = 0xFFFF0000
	MsgClientLogon_ProtocolVerMinorMask                                 uint32 = 0xFFFF
	MsgClientLogon_ProtocolVerMinorMinGameServers                       uint16 = 4
	MsgClientLogon_ProtocolVerMinorMinForSupportingEMsgMulti            uint16 = 12
	MsgClientLogon_ProtocolVerMinorMinForSupportingEMsgClientEncryptPct uint16 = 14
	MsgClientLogon_ProtocolVerMinorMinForExtendedMsgHdr                 uint16 = 17
	MsgClientLogon_ProtocolVerMinorMinForCellId                         uint16 = 18
	MsgClientLogon_ProtocolVerMinorMinForSessionIDLast                  uint16 = 19
	MsgClientLogon_ProtocolVerMinorMinForServerAvailablityMsgs          uint16 = 24
	MsgClientLogon_ProtocolVerMinorMinClients                           uint16 = 25
	MsgClientLogon_ProtocolVerMinorMinForOSType                         uint16 = 26
	MsgClientLogon_ProtocolVerMinorMinForCegApplyPESig                  uint16 = 27
	MsgClientLogon_ProtocolVerMinorMinForMarketingMessages2             uint16 = 27
	MsgClientLogon_ProtocolVerMinorMinForAnyProtoBufMessages            uint16 = 28
	MsgClientLogon_ProtocolVerMinorMinForProtoBufLoggedOffMessage       uint16 = 28
	MsgClientLogon_ProtocolVerMinorMinForProtoBufMultiMessages          uint16 = 28
	MsgClientLogon_ProtocolVerMinorMinForSendingProtocolToUFS           uint16 = 30
	MsgClientLogon_ProtocolVerMinorMinForMachineAuth                    uint16 = 33
	MsgClientLogon_ProtocolVerMinorMinForSessionIDLastAnon              uint16 = 36
	MsgClientLogon_ProtocolVerMinorMinForEnhancedAppList                uint16 = 40
	MsgClientLogon_ProtocolVerMinorMinForSteamGuardNotificationUI       uint16 = 41
	MsgClientLogon_ProtocolVerMinorMinForProtoBufServiceModuleCalls     uint16 = 42
	MsgClientLogon_ProtocolVerMinorMinForGzipMultiMessages              uint16 = 43
	MsgClientLogon_ProtocolVerMinorMinForNewVoiceCallAuthorize          uint16 = 44
	MsgClientLogon_ProtocolVerMinorMinForClientInstanceIDs              uint16 = 44
)

type MsgClientLogon struct {
}

func NewMsgClientLogon() *MsgClientLogon {
	return &MsgClientLogon{}
}

func (d *MsgClientLogon) GetEMsg() EMsg {
	return EMsg_ClientLogon
}

func (d *MsgClientLogon) Serialize(w io.Writer) error {
	var err error
	return err
}

func (d *MsgClientLogon) Deserialize(r io.Reader) error {
	var err error
	return err
}

type MsgClientVACBanStatus struct {
	NumBans uint32
}

func NewMsgClientVACBanStatus() *MsgClientVACBanStatus {
	return &MsgClientVACBanStatus{}
}

func (d *MsgClientVACBanStatus) GetEMsg() EMsg {
	return EMsg_ClientVACBanStatus
}

func (d *MsgClientVACBanStatus) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.NumBans)
	return err
}

func (d *MsgClientVACBanStatus) Deserialize(r io.Reader) error {
	var err error
	d.NumBans, err = rwu.ReadUint32(r)
	return err
}

type MsgClientAppUsageEvent struct {
	AppUsageEvent EAppUsageEvent
	GameID        uint64
	Offline       uint16
}

func NewMsgClientAppUsageEvent() *MsgClientAppUsageEvent {
	return &MsgClientAppUsageEvent{}
}

func (d *MsgClientAppUsageEvent) GetEMsg() EMsg {
	return EMsg_ClientAppUsageEvent
}

func (d *MsgClientAppUsageEvent) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.AppUsageEvent)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.GameID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Offline)
	return err
}

func (d *MsgClientAppUsageEvent) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.AppUsageEvent = EAppUsageEvent(t0)
	d.GameID, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.Offline, err = rwu.ReadUint16(r)
	return err
}

type MsgClientEmailAddrInfo struct {
	PasswordStrength           uint32
	FlagsAccountSecurityPolicy uint32
	Validated                  bool
}

func NewMsgClientEmailAddrInfo() *MsgClientEmailAddrInfo {
	return &MsgClientEmailAddrInfo{}
}

func (d *MsgClientEmailAddrInfo) GetEMsg() EMsg {
	return EMsg_ClientEmailAddrInfo
}

func (d *MsgClientEmailAddrInfo) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.PasswordStrength)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.FlagsAccountSecurityPolicy)
	if err != nil {
		return err
	}
	err = rwu.WriteBool(w, d.Validated)
	return err
}

func (d *MsgClientEmailAddrInfo) Deserialize(r io.Reader) error {
	var err error
	d.PasswordStrength, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.FlagsAccountSecurityPolicy, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.Validated, err = rwu.ReadBool(r)
	return err
}

type MsgClientUpdateGuestPassesList struct {
	Result                   EResult
	CountGuestPassesToGive   int32
	CountGuestPassesToRedeem int32
}

func NewMsgClientUpdateGuestPassesList() *MsgClientUpdateGuestPassesList {
	return &MsgClientUpdateGuestPassesList{}
}

func (d *MsgClientUpdateGuestPassesList) GetEMsg() EMsg {
	return EMsg_ClientUpdateGuestPassesList
}

func (d *MsgClientUpdateGuestPassesList) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.CountGuestPassesToGive)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.CountGuestPassesToRedeem)
	return err
}

func (d *MsgClientUpdateGuestPassesList) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	d.CountGuestPassesToGive, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.CountGuestPassesToRedeem, err = rwu.ReadInt32(r)
	return err
}

type MsgClientRequestedClientStats struct {
	CountStats int32
}

func NewMsgClientRequestedClientStats() *MsgClientRequestedClientStats {
	return &MsgClientRequestedClientStats{}
}

func (d *MsgClientRequestedClientStats) GetEMsg() EMsg {
	return EMsg_ClientRequestedClientStats
}

func (d *MsgClientRequestedClientStats) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.CountStats)
	return err
}

func (d *MsgClientRequestedClientStats) Deserialize(r io.Reader) error {
	var err error
	d.CountStats, err = rwu.ReadInt32(r)
	return err
}

type MsgClientP2PIntroducerMessage struct {
	SteamID     steamid.SteamId
	RoutingType EIntroducerRouting
	Data        []uint8
	DataLen     uint32
}

func NewMsgClientP2PIntroducerMessage() *MsgClientP2PIntroducerMessage {
	return &MsgClientP2PIntroducerMessage{
		Data: make([]uint8, 1450, 1450),
	}
}

func (d *MsgClientP2PIntroducerMessage) GetEMsg() EMsg {
	return EMsg_ClientP2PIntroducerMessage
}

func (d *MsgClientP2PIntroducerMessage) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.RoutingType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Data)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.DataLen)
	return err
}

func (d *MsgClientP2PIntroducerMessage) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamID = steamid.SteamId(t0)
	t1, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.RoutingType = EIntroducerRouting(t1)
	err = binary.Read(r, binary.LittleEndian, d.Data)
	if err != nil {
		return err
	}
	d.DataLen, err = rwu.ReadUint32(r)
	return err
}

type MsgClientOGSBeginSession struct {
	AccountType uint8
	AccountId   steamid.SteamId
	AppId       uint32
	TimeStarted uint32
}

func NewMsgClientOGSBeginSession() *MsgClientOGSBeginSession {
	return &MsgClientOGSBeginSession{}
}

func (d *MsgClientOGSBeginSession) GetEMsg() EMsg {
	return EMsg_ClientOGSBeginSession
}

func (d *MsgClientOGSBeginSession) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.AccountType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.AccountId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.AppId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.TimeStarted)
	return err
}

func (d *MsgClientOGSBeginSession) Deserialize(r io.Reader) error {
	var err error
	d.AccountType, err = rwu.ReadUint8(r)
	if err != nil {
		return err
	}
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.AccountId = steamid.SteamId(t0)
	d.AppId, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.TimeStarted, err = rwu.ReadUint32(r)
	return err
}

type MsgClientOGSBeginSessionResponse struct {
	Result            EResult
	CollectingAny     bool
	CollectingDetails bool
	SessionId         uint64
}

func NewMsgClientOGSBeginSessionResponse() *MsgClientOGSBeginSessionResponse {
	return &MsgClientOGSBeginSessionResponse{}
}

func (d *MsgClientOGSBeginSessionResponse) GetEMsg() EMsg {
	return EMsg_ClientOGSBeginSessionResponse
}

func (d *MsgClientOGSBeginSessionResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = rwu.WriteBool(w, d.CollectingAny)
	if err != nil {
		return err
	}
	err = rwu.WriteBool(w, d.CollectingDetails)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SessionId)
	return err
}

func (d *MsgClientOGSBeginSessionResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	d.CollectingAny, err = rwu.ReadBool(r)
	if err != nil {
		return err
	}
	d.CollectingDetails, err = rwu.ReadBool(r)
	if err != nil {
		return err
	}
	d.SessionId, err = rwu.ReadUint64(r)
	return err
}

type MsgClientOGSEndSession struct {
	SessionId       uint64
	TimeEnded       uint32
	ReasonCode      int32
	CountAttributes int32
}

func NewMsgClientOGSEndSession() *MsgClientOGSEndSession {
	return &MsgClientOGSEndSession{}
}

func (d *MsgClientOGSEndSession) GetEMsg() EMsg {
	return EMsg_ClientOGSEndSession
}

func (d *MsgClientOGSEndSession) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SessionId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.TimeEnded)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ReasonCode)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.CountAttributes)
	return err
}

func (d *MsgClientOGSEndSession) Deserialize(r io.Reader) error {
	var err error
	d.SessionId, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.TimeEnded, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.ReasonCode, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.CountAttributes, err = rwu.ReadInt32(r)
	return err
}

type MsgClientOGSEndSessionResponse struct {
	Result EResult
}

func NewMsgClientOGSEndSessionResponse() *MsgClientOGSEndSessionResponse {
	return &MsgClientOGSEndSessionResponse{}
}

func (d *MsgClientOGSEndSessionResponse) GetEMsg() EMsg {
	return EMsg_ClientOGSEndSessionResponse
}

func (d *MsgClientOGSEndSessionResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	return err
}

func (d *MsgClientOGSEndSessionResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	d.Result = EResult(t0)
	return err
}

type MsgClientOGSWriteRow struct {
	SessionId       uint64
	CountAttributes int32
}

func NewMsgClientOGSWriteRow() *MsgClientOGSWriteRow {
	return &MsgClientOGSWriteRow{}
}

func (d *MsgClientOGSWriteRow) GetEMsg() EMsg {
	return EMsg_ClientOGSWriteRow
}

func (d *MsgClientOGSWriteRow) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SessionId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.CountAttributes)
	return err
}

func (d *MsgClientOGSWriteRow) Deserialize(r io.Reader) error {
	var err error
	d.SessionId, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.CountAttributes, err = rwu.ReadInt32(r)
	return err
}

type MsgClientGetFriendsWhoPlayGame struct {
	GameId uint64
}

func NewMsgClientGetFriendsWhoPlayGame() *MsgClientGetFriendsWhoPlayGame {
	return &MsgClientGetFriendsWhoPlayGame{}
}

func (d *MsgClientGetFriendsWhoPlayGame) GetEMsg() EMsg {
	return EMsg_ClientGetFriendsWhoPlayGame
}

func (d *MsgClientGetFriendsWhoPlayGame) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.GameId)
	return err
}

func (d *MsgClientGetFriendsWhoPlayGame) Deserialize(r io.Reader) error {
	var err error
	d.GameId, err = rwu.ReadUint64(r)
	return err
}

type MsgClientGetFriendsWhoPlayGameResponse struct {
	Result       EResult
	GameId       uint64
	CountFriends uint32
}

func NewMsgClientGetFriendsWhoPlayGameResponse() *MsgClientGetFriendsWhoPlayGameResponse {
	return &MsgClientGetFriendsWhoPlayGameResponse{}
}

func (d *MsgClientGetFriendsWhoPlayGameResponse) GetEMsg() EMsg {
	return EMsg_ClientGetFriendsWhoPlayGameResponse
}

func (d *MsgClientGetFriendsWhoPlayGameResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.GameId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.CountFriends)
	return err
}

func (d *MsgClientGetFriendsWhoPlayGameResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	d.GameId, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.CountFriends, err = rwu.ReadUint32(r)
	return err
}

type MsgGSPerformHardwareSurvey struct {
	Flags uint32
}

func NewMsgGSPerformHardwareSurvey() *MsgGSPerformHardwareSurvey {
	return &MsgGSPerformHardwareSurvey{}
}

func (d *MsgGSPerformHardwareSurvey) GetEMsg() EMsg {
	return EMsg_GSPerformHardwareSurvey
}

func (d *MsgGSPerformHardwareSurvey) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Flags)
	return err
}

func (d *MsgGSPerformHardwareSurvey) Deserialize(r io.Reader) error {
	var err error
	d.Flags, err = rwu.ReadUint32(r)
	return err
}

type MsgGSGetPlayStatsResponse struct {
	Result                EResult
	Rank                  int32
	LifetimeConnects      uint32
	LifetimeMinutesPlayed uint32
}

func NewMsgGSGetPlayStatsResponse() *MsgGSGetPlayStatsResponse {
	return &MsgGSGetPlayStatsResponse{}
}

func (d *MsgGSGetPlayStatsResponse) GetEMsg() EMsg {
	return EMsg_GSGetPlayStatsResponse
}

func (d *MsgGSGetPlayStatsResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Rank)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.LifetimeConnects)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.LifetimeMinutesPlayed)
	return err
}

func (d *MsgGSGetPlayStatsResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	d.Rank, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.LifetimeConnects, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.LifetimeMinutesPlayed, err = rwu.ReadUint32(r)
	return err
}

type MsgGSGetReputationResponse struct {
	Result          EResult
	ReputationScore uint32
	Banned          bool
	BannedIp        uint32
	BannedPort      uint16
	BannedGameId    uint64
	TimeBanExpires  uint32
}

func NewMsgGSGetReputationResponse() *MsgGSGetReputationResponse {
	return &MsgGSGetReputationResponse{}
}

func (d *MsgGSGetReputationResponse) GetEMsg() EMsg {
	return EMsg_GSGetReputationResponse
}

func (d *MsgGSGetReputationResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ReputationScore)
	if err != nil {
		return err
	}
	err = rwu.WriteBool(w, d.Banned)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.BannedIp)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.BannedPort)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.BannedGameId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.TimeBanExpires)
	return err
}

func (d *MsgGSGetReputationResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	d.ReputationScore, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.Banned, err = rwu.ReadBool(r)
	if err != nil {
		return err
	}
	d.BannedIp, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.BannedPort, err = rwu.ReadUint16(r)
	if err != nil {
		return err
	}
	d.BannedGameId, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.TimeBanExpires, err = rwu.ReadUint32(r)
	return err
}

type MsgGSDeny struct {
	SteamId    steamid.SteamId
	DenyReason EDenyReason
}

func NewMsgGSDeny() *MsgGSDeny {
	return &MsgGSDeny{}
}

func (d *MsgGSDeny) GetEMsg() EMsg {
	return EMsg_GSDeny
}

func (d *MsgGSDeny) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.DenyReason)
	return err
}

func (d *MsgGSDeny) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamId = steamid.SteamId(t0)
	t1, err := rwu.ReadInt32(r)
	d.DenyReason = EDenyReason(t1)
	return err
}

type MsgGSApprove struct {
	SteamId steamid.SteamId
}

func NewMsgGSApprove() *MsgGSApprove {
	return &MsgGSApprove{}
}

func (d *MsgGSApprove) GetEMsg() EMsg {
	return EMsg_GSApprove
}

func (d *MsgGSApprove) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamId)
	return err
}

func (d *MsgGSApprove) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamId = steamid.SteamId(t0)
	return err
}

type MsgGSKick struct {
	SteamId          steamid.SteamId
	DenyReason       EDenyReason
	WaitTilMapChange int32
}

func NewMsgGSKick() *MsgGSKick {
	return &MsgGSKick{}
}

func (d *MsgGSKick) GetEMsg() EMsg {
	return EMsg_GSKick
}

func (d *MsgGSKick) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.DenyReason)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.WaitTilMapChange)
	return err
}

func (d *MsgGSKick) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamId = steamid.SteamId(t0)
	t1, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.DenyReason = EDenyReason(t1)
	d.WaitTilMapChange, err = rwu.ReadInt32(r)
	return err
}

type MsgGSGetUserGroupStatus struct {
	SteamIdUser  steamid.SteamId
	SteamIdGroup steamid.SteamId
}

func NewMsgGSGetUserGroupStatus() *MsgGSGetUserGroupStatus {
	return &MsgGSGetUserGroupStatus{}
}

func (d *MsgGSGetUserGroupStatus) GetEMsg() EMsg {
	return EMsg_GSGetUserGroupStatus
}

func (d *MsgGSGetUserGroupStatus) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdUser)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdGroup)
	return err
}

func (d *MsgGSGetUserGroupStatus) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdUser = steamid.SteamId(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdGroup = steamid.SteamId(t1)
	return err
}

type MsgGSGetUserGroupStatusResponse struct {
	SteamIdUser      steamid.SteamId
	SteamIdGroup     steamid.SteamId
	ClanRelationship EClanRelationship
	ClanRank         EClanRank
}

func NewMsgGSGetUserGroupStatusResponse() *MsgGSGetUserGroupStatusResponse {
	return &MsgGSGetUserGroupStatusResponse{}
}

func (d *MsgGSGetUserGroupStatusResponse) GetEMsg() EMsg {
	return EMsg_GSGetUserGroupStatusResponse
}

func (d *MsgGSGetUserGroupStatusResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdUser)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdGroup)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ClanRelationship)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ClanRank)
	return err
}

func (d *MsgGSGetUserGroupStatusResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdUser = steamid.SteamId(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdGroup = steamid.SteamId(t1)
	t2, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.ClanRelationship = EClanRelationship(t2)
	t3, err := rwu.ReadInt32(r)
	d.ClanRank = EClanRank(t3)
	return err
}

type MsgClientJoinChat struct {
	SteamIdChat    steamid.SteamId
	IsVoiceSpeaker bool
}

func NewMsgClientJoinChat() *MsgClientJoinChat {
	return &MsgClientJoinChat{}
}

func (d *MsgClientJoinChat) GetEMsg() EMsg {
	return EMsg_ClientJoinChat
}

func (d *MsgClientJoinChat) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChat)
	if err != nil {
		return err
	}
	err = rwu.WriteBool(w, d.IsVoiceSpeaker)
	return err
}

func (d *MsgClientJoinChat) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChat = steamid.SteamId(t0)
	d.IsVoiceSpeaker, err = rwu.ReadBool(r)
	return err
}

type MsgClientChatEnter struct {
	SteamIdChat   steamid.SteamId
	SteamIdFriend steamid.SteamId
	ChatRoomType  EChatRoomType
	SteamIdOwner  steamid.SteamId
	SteamIdClan   steamid.SteamId
	ChatFlags     uint8
	EnterResponse EChatRoomEnterResponse
	NumMembers    int32
}

func NewMsgClientChatEnter() *MsgClientChatEnter {
	return &MsgClientChatEnter{}
}

func (d *MsgClientChatEnter) GetEMsg() EMsg {
	return EMsg_ClientChatEnter
}

func (d *MsgClientChatEnter) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChat)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdFriend)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ChatRoomType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdOwner)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdClan)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ChatFlags)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.EnterResponse)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.NumMembers)
	return err
}

func (d *MsgClientChatEnter) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChat = steamid.SteamId(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdFriend = steamid.SteamId(t1)
	t2, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.ChatRoomType = EChatRoomType(t2)
	t3, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdOwner = steamid.SteamId(t3)
	t4, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdClan = steamid.SteamId(t4)
	d.ChatFlags, err = rwu.ReadUint8(r)
	if err != nil {
		return err
	}
	t5, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.EnterResponse = EChatRoomEnterResponse(t5)
	d.NumMembers, err = rwu.ReadInt32(r)
	return err
}

type MsgClientChatMsg struct {
	SteamIdChatter  steamid.SteamId
	SteamIdChatRoom steamid.SteamId
	ChatMsgType     EChatEntryType
}

func NewMsgClientChatMsg() *MsgClientChatMsg {
	return &MsgClientChatMsg{}
}

func (d *MsgClientChatMsg) GetEMsg() EMsg {
	return EMsg_ClientChatMsg
}

func (d *MsgClientChatMsg) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChatter)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChatRoom)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ChatMsgType)
	return err
}

func (d *MsgClientChatMsg) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChatter = steamid.SteamId(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChatRoom = steamid.SteamId(t1)
	t2, err := rwu.ReadInt32(r)
	d.ChatMsgType = EChatEntryType(t2)
	return err
}

type MsgClientChatMemberInfo struct {
	SteamIdChat steamid.SteamId
	Type        EChatInfoType
}

func NewMsgClientChatMemberInfo() *MsgClientChatMemberInfo {
	return &MsgClientChatMemberInfo{}
}

func (d *MsgClientChatMemberInfo) GetEMsg() EMsg {
	return EMsg_ClientChatMemberInfo
}

func (d *MsgClientChatMemberInfo) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChat)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Type)
	return err
}

func (d *MsgClientChatMemberInfo) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChat = steamid.SteamId(t0)
	t1, err := rwu.ReadInt32(r)
	d.Type = EChatInfoType(t1)
	return err
}

type MsgClientChatAction struct {
	SteamIdChat        steamid.SteamId
	SteamIdUserToActOn steamid.SteamId
	ChatAction         EChatAction
}

func NewMsgClientChatAction() *MsgClientChatAction {
	return &MsgClientChatAction{}
}

func (d *MsgClientChatAction) GetEMsg() EMsg {
	return EMsg_ClientChatAction
}

func (d *MsgClientChatAction) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChat)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdUserToActOn)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ChatAction)
	return err
}

func (d *MsgClientChatAction) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChat = steamid.SteamId(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdUserToActOn = steamid.SteamId(t1)
	t2, err := rwu.ReadInt32(r)
	d.ChatAction = EChatAction(t2)
	return err
}

type MsgClientChatActionResult struct {
	SteamIdChat        steamid.SteamId
	SteamIdUserActedOn steamid.SteamId
	ChatAction         EChatAction
	ActionResult       EChatActionResult
}

func NewMsgClientChatActionResult() *MsgClientChatActionResult {
	return &MsgClientChatActionResult{}
}

func (d *MsgClientChatActionResult) GetEMsg() EMsg {
	return EMsg_ClientChatActionResult
}

func (d *MsgClientChatActionResult) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChat)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdUserActedOn)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ChatAction)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ActionResult)
	return err
}

func (d *MsgClientChatActionResult) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChat = steamid.SteamId(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdUserActedOn = steamid.SteamId(t1)
	t2, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.ChatAction = EChatAction(t2)
	t3, err := rwu.ReadInt32(r)
	d.ActionResult = EChatActionResult(t3)
	return err
}

type MsgClientChatRoomInfo struct {
	SteamIdChat steamid.SteamId
	Type        EChatInfoType
}

func NewMsgClientChatRoomInfo() *MsgClientChatRoomInfo {
	return &MsgClientChatRoomInfo{}
}

func (d *MsgClientChatRoomInfo) GetEMsg() EMsg {
	return EMsg_ClientChatRoomInfo
}

func (d *MsgClientChatRoomInfo) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChat)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Type)
	return err
}

func (d *MsgClientChatRoomInfo) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChat = steamid.SteamId(t0)
	t1, err := rwu.ReadInt32(r)
	d.Type = EChatInfoType(t1)
	return err
}

type MsgClientSetIgnoreFriend struct {
	MySteamId     steamid.SteamId
	SteamIdFriend steamid.SteamId
	Ignore        uint8
}

func NewMsgClientSetIgnoreFriend() *MsgClientSetIgnoreFriend {
	return &MsgClientSetIgnoreFriend{}
}

func (d *MsgClientSetIgnoreFriend) GetEMsg() EMsg {
	return EMsg_ClientSetIgnoreFriend
}

func (d *MsgClientSetIgnoreFriend) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.MySteamId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdFriend)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Ignore)
	return err
}

func (d *MsgClientSetIgnoreFriend) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.MySteamId = steamid.SteamId(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdFriend = steamid.SteamId(t1)
	d.Ignore, err = rwu.ReadUint8(r)
	return err
}

type MsgClientSetIgnoreFriendResponse struct {
	FriendId steamid.SteamId
	Result   EResult
}

func NewMsgClientSetIgnoreFriendResponse() *MsgClientSetIgnoreFriendResponse {
	return &MsgClientSetIgnoreFriendResponse{}
}

func (d *MsgClientSetIgnoreFriendResponse) GetEMsg() EMsg {
	return EMsg_ClientSetIgnoreFriendResponse
}

func (d *MsgClientSetIgnoreFriendResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.FriendId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Result)
	return err
}

func (d *MsgClientSetIgnoreFriendResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.FriendId = steamid.SteamId(t0)
	t1, err := rwu.ReadInt32(r)
	d.Result = EResult(t1)
	return err
}

type MsgClientLoggedOff struct {
	Result              EResult
	SecMinReconnectHint int32
	SecMaxReconnectHint int32
}

func NewMsgClientLoggedOff() *MsgClientLoggedOff {
	return &MsgClientLoggedOff{}
}

func (d *MsgClientLoggedOff) GetEMsg() EMsg {
	return EMsg_ClientLoggedOff
}

func (d *MsgClientLoggedOff) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SecMinReconnectHint)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SecMaxReconnectHint)
	return err
}

func (d *MsgClientLoggedOff) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	d.SecMinReconnectHint, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.SecMaxReconnectHint, err = rwu.ReadInt32(r)
	return err
}

type MsgClientLogOnResponse struct {
	Result                    EResult
	OutOfGameHeartbeatRateSec int32
	InGameHeartbeatRateSec    int32
	ClientSuppliedSteamId     steamid.SteamId
	IpPublic                  uint32
	ServerRealTime            uint32
}

func NewMsgClientLogOnResponse() *MsgClientLogOnResponse {
	return &MsgClientLogOnResponse{}
}

func (d *MsgClientLogOnResponse) GetEMsg() EMsg {
	return EMsg_ClientLogOnResponse
}

func (d *MsgClientLogOnResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.OutOfGameHeartbeatRateSec)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.InGameHeartbeatRateSec)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ClientSuppliedSteamId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.IpPublic)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ServerRealTime)
	return err
}

func (d *MsgClientLogOnResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	d.OutOfGameHeartbeatRateSec, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.InGameHeartbeatRateSec, err = rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.ClientSuppliedSteamId = steamid.SteamId(t1)
	d.IpPublic, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.ServerRealTime, err = rwu.ReadUint32(r)
	return err
}

type MsgClientServerUnavailable struct {
	JobidSent              uint64
	EMsgSent               uint32
	EServerTypeUnavailable EServerType
}

func NewMsgClientServerUnavailable() *MsgClientServerUnavailable {
	return &MsgClientServerUnavailable{}
}

func (d *MsgClientServerUnavailable) GetEMsg() EMsg {
	return EMsg_ClientServerUnavailable
}

func (d *MsgClientServerUnavailable) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.JobidSent)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.EMsgSent)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.EServerTypeUnavailable)
	return err
}

func (d *MsgClientServerUnavailable) Deserialize(r io.Reader) error {
	var err error
	d.JobidSent, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.EMsgSent, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	t0, err := rwu.ReadInt32(r)
	d.EServerTypeUnavailable = EServerType(t0)
	return err
}

type MsgClientCreateChat struct {
	ChatRoomType      EChatRoomType
	GameId            uint64
	SteamIdClan       steamid.SteamId
	PermissionOfficer EChatPermission
	PermissionMember  EChatPermission
	PermissionAll     EChatPermission
	MembersMax        uint32
	ChatFlags         uint8
	SteamIdFriendChat steamid.SteamId
	SteamIdInvited    steamid.SteamId
}

func NewMsgClientCreateChat() *MsgClientCreateChat {
	return &MsgClientCreateChat{}
}

func (d *MsgClientCreateChat) GetEMsg() EMsg {
	return EMsg_ClientCreateChat
}

func (d *MsgClientCreateChat) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.ChatRoomType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.GameId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdClan)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.PermissionOfficer)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.PermissionMember)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.PermissionAll)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.MembersMax)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ChatFlags)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdFriendChat)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdInvited)
	return err
}

func (d *MsgClientCreateChat) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.ChatRoomType = EChatRoomType(t0)
	d.GameId, err = rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdClan = steamid.SteamId(t1)
	t2, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.PermissionOfficer = EChatPermission(t2)
	t3, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.PermissionMember = EChatPermission(t3)
	t4, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.PermissionAll = EChatPermission(t4)
	d.MembersMax, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.ChatFlags, err = rwu.ReadUint8(r)
	if err != nil {
		return err
	}
	t5, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdFriendChat = steamid.SteamId(t5)
	t6, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdInvited = steamid.SteamId(t6)
	return err
}

type MsgClientCreateChatResponse struct {
	Result            EResult
	SteamIdChat       steamid.SteamId
	ChatRoomType      EChatRoomType
	SteamIdFriendChat steamid.SteamId
}

func NewMsgClientCreateChatResponse() *MsgClientCreateChatResponse {
	return &MsgClientCreateChatResponse{}
}

func (d *MsgClientCreateChatResponse) GetEMsg() EMsg {
	return EMsg_ClientCreateChatResponse
}

func (d *MsgClientCreateChatResponse) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.Result)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdChat)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.ChatRoomType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.SteamIdFriendChat)
	return err
}

func (d *MsgClientCreateChatResponse) Deserialize(r io.Reader) error {
	var err error
	t0, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.Result = EResult(t0)
	t1, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdChat = steamid.SteamId(t1)
	t2, err := rwu.ReadInt32(r)
	if err != nil {
		return err
	}
	d.ChatRoomType = EChatRoomType(t2)
	t3, err := rwu.ReadUint64(r)
	if err != nil {
		return err
	}
	d.SteamIdFriendChat = steamid.SteamId(t3)
	return err
}

type MsgClientMarketingMessageUpdate2 struct {
	MarketingMessageUpdateTime uint32
	Count                      uint32
}

func NewMsgClientMarketingMessageUpdate2() *MsgClientMarketingMessageUpdate2 {
	return &MsgClientMarketingMessageUpdate2{}
}

func (d *MsgClientMarketingMessageUpdate2) GetEMsg() EMsg {
	return EMsg_ClientMarketingMessageUpdate2
}

func (d *MsgClientMarketingMessageUpdate2) Serialize(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.LittleEndian, d.MarketingMessageUpdateTime)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Count)
	return err
}

func (d *MsgClientMarketingMessageUpdate2) Deserialize(r io.Reader) error {
	var err error
	d.MarketingMessageUpdateTime, err = rwu.ReadUint32(r)
	if err != nil {
		return err
	}
	d.Count, err = rwu.ReadUint32(r)
	return err
}
