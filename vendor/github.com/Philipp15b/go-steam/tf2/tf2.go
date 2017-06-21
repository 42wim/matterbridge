/*
Provides access to TF2 Game Coordinator functionality.
*/
package tf2

import (
	"github.com/Philipp15b/go-steam"
	. "github.com/Philipp15b/go-steam/protocol/gamecoordinator"
	. "github.com/Philipp15b/go-steam/tf2/protocol"
	"github.com/Philipp15b/go-steam/tf2/protocol/protobuf"
)

const AppId = 440

// To use any methods of this, you'll need to SetPlaying(true) and wait for
// the GCReadyEvent.
type TF2 struct {
	client *steam.Client
}

// Creates a new TF2 instance and registers it as a packet handler
func New(client *steam.Client) *TF2 {
	t := &TF2{client}
	client.GC.RegisterPacketHandler(t)
	return t
}

func (t *TF2) SetPlaying(playing bool) {
	if playing {
		t.client.GC.SetGamesPlayed(AppId)
	} else {
		t.client.GC.SetGamesPlayed()
	}
}

func (t *TF2) SetItemPosition(itemId, position uint64) {
	t.client.GC.Write(NewGCMsg(AppId, uint32(protobuf.EGCItemMsg_k_EMsgGCSetSingleItemPosition), &MsgGCSetItemPosition{
		itemId, position,
	}))
}

// recipe -2 = wildcard
func (t *TF2) CraftItems(items []uint64, recipe int16) {
	t.client.GC.Write(NewGCMsg(AppId, uint32(protobuf.EGCItemMsg_k_EMsgGCCraft), &MsgGCCraft{
		Recipe: recipe,
		Items:  items,
	}))
}

func (t *TF2) DeleteItem(itemId uint64) {
	t.client.GC.Write(NewGCMsg(AppId, uint32(protobuf.EGCItemMsg_k_EMsgGCDelete), &MsgGCDeleteItem{itemId}))
}

func (t *TF2) NameItem(toolId, target uint64, name string) {
	t.client.GC.Write(NewGCMsg(AppId, uint32(protobuf.EGCItemMsg_k_EMsgGCNameItem), &MsgGCNameItem{
		toolId, target, name,
	}))
}

type GCReadyEvent struct{}

func (t *TF2) HandleGCPacket(packet *GCPacket) {
	if packet.AppId != AppId {
		return
	}
	switch protobuf.EGCBaseClientMsg(packet.MsgType) {
	case protobuf.EGCBaseClientMsg_k_EMsgGCClientWelcome:
		t.handleWelcome(packet)
	}
}

func (t *TF2) handleWelcome(packet *GCPacket) {
	// the packet's body is pretty useless
	t.client.Emit(&GCReadyEvent{})
}
