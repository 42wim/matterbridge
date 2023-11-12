package sctp

import (
	"errors"
	"fmt"
)

/*
chunkHeartbeat represents an SCTP Chunk of type HEARTBEAT

An endpoint should send this chunk to its peer endpoint to probe the
reachability of a particular destination transport address defined in
the present association.

The parameter field contains the Heartbeat Information, which is a
variable-length opaque data structure understood only by the sender.


 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|   Type = 4    | Chunk  Flags  |      Heartbeat Length         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|            Heartbeat Information TLV (Variable-Length)        |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Defined as a variable-length parameter using the format described
in Section 3.2.1, i.e.:

Variable Parameters                  Status     Type Value
-------------------------------------------------------------
heartbeat Info                       Mandatory   1
*/
type chunkHeartbeat struct {
	chunkHeader
	params []param
}

var (
	errChunkTypeNotHeartbeat      = errors.New("ChunkType is not of type HEARTBEAT")
	errHeartbeatNotLongEnoughInfo = errors.New("heartbeat is not long enough to contain Heartbeat Info")
	errParseParamTypeFailed       = errors.New("failed to parse param type")
	errHeartbeatParam             = errors.New("heartbeat should only have HEARTBEAT param")
	errHeartbeatChunkUnmarshal    = errors.New("failed unmarshalling param in Heartbeat Chunk")
)

func (h *chunkHeartbeat) unmarshal(raw []byte) error {
	if err := h.chunkHeader.unmarshal(raw); err != nil {
		return err
	} else if h.typ != ctHeartbeat {
		return fmt.Errorf("%w: actually is %s", errChunkTypeNotHeartbeat, h.typ.String())
	}

	if len(raw) <= chunkHeaderSize {
		return fmt.Errorf("%w: %d", errHeartbeatNotLongEnoughInfo, len(raw))
	}

	pType, err := parseParamType(raw[chunkHeaderSize:])
	if err != nil {
		return fmt.Errorf("%w: %v", errParseParamTypeFailed, err)
	}
	if pType != heartbeatInfo {
		return fmt.Errorf("%w: instead have %s", errHeartbeatParam, pType.String())
	}

	p, err := buildParam(pType, raw[chunkHeaderSize:])
	if err != nil {
		return fmt.Errorf("%w: %v", errHeartbeatChunkUnmarshal, err)
	}
	h.params = append(h.params, p)

	return nil
}

func (h *chunkHeartbeat) Marshal() ([]byte, error) {
	return nil, errUnimplemented
}

func (h *chunkHeartbeat) check() (abort bool, err error) {
	return false, nil
}
