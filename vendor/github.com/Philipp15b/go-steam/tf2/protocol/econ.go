package protocol

import (
	"encoding/binary"
	"io"
)

type MsgGCSetItemPosition struct {
	AssetId, Position uint64
}

func (m *MsgGCSetItemPosition) Serialize(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, m)
}

type MsgGCCraft struct {
	Recipe   int16 // -2 = wildcard
	numItems int16
	Items    []uint64
}

func (m *MsgGCCraft) Serialize(w io.Writer) error {
	m.numItems = int16(len(m.Items))
	return binary.Write(w, binary.LittleEndian, m)
}

type MsgGCDeleteItem struct {
	ItemId uint64
}

func (m *MsgGCDeleteItem) Serialize(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, m.ItemId)
}

type MsgGCNameItem struct {
	Tool, Target uint64
	Name         string
}

func (m *MsgGCNameItem) Serialize(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, m.Tool)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, m.Target)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(m.Name))
	return err
}
