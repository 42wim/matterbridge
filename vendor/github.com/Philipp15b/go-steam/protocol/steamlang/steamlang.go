/*
Contains code generated from SteamKit's SteamLanguage data.
*/
package steamlang

const (
	ProtoMask uint32 = 0x80000000
	EMsgMask         = ^ProtoMask
)

func NewEMsg(e uint32) EMsg {
	return EMsg(e & EMsgMask)
}

func IsProto(e uint32) bool {
	return e&ProtoMask > 0
}
