package ldap

import (
	"bytes"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

type debugging struct {
	logger mlog.LoggerIFace
	levels []mlog.Level
}

// Enable controls debugging mode.
func (debug *debugging) Enable(logger mlog.LoggerIFace, levels ...mlog.Level) {
	*debug = debugging{
		logger: logger,
		levels: levels,
	}
}

func (debug debugging) Enabled() bool {
	return debug.logger != nil
}

// Log writes debug output.
func (debug debugging) Log(msg string, fields ...mlog.Field) {
	if debug.Enabled() {
		debug.logger.LogM(debug.levels, msg, fields...)
	}
}

type Packet ber.Packet

func (p Packet) LogClone() any {
	bp := ber.Packet(p)
	var b bytes.Buffer
	ber.WritePacket(&b, &bp)
	return b.String()

}

func PacketToField(packet *ber.Packet) mlog.Field {
	if packet == nil {
		return mlog.Any("packet", nil)
	}
	return mlog.Any("packet", Packet(*packet))
}
