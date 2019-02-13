package bwhatsapp

import (
	"github.com/42wim/matterbridge/bridge/config"

	"github.com/Rhymen/go-whatsapp"
)

/*
Implement handling messages coming from the bridge to WhatsApp
*/

// Send a message from the bridge to WhatsApp
// Required implementation of the Bridger interface
// https://github.com/42wim/matterbridge/blob/2cfd880cdb0df29771bf8f31df8d990ab897889d/bridge/bridge.go#L11-L16
func (b *Bwhatsapp) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// msg.Channel target group name
	// msg.Username empty
	// msg.UserID a weird string , probably slack user id
	// msg.Avatar has a nice image
	// msg.Timestamp has a nice timestamp with loc(ation) / timezone
	// msg.ID empty, // TODO why empty?!

	text := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			// Id: "", // TODO id
			// TODO Timestamp
			RemoteJid: msg.Channel, // which equals to group id

		},
		Text: msg.Username + msg.Text,
	}

	// TODO adapt gitter code for edits, delete and some extra commands
	//roomID := b.getRoomID(msg.Channel)
	//if roomID == "" {
	//	b.Log.Errorf("Could not find roomID for %v", msg.Channel)
	//	return "", nil
	//}
	//
	//// Delete message
	//if msg.Event == config.EventMsgDelete {
	//	if msg.ID == "" {
	//		return "", nil
	//	}
	//	// gitter has no delete message api so we edit message to ""
	//	_, err := b.c.UpdateMessage(roomID, msg.ID, "")
	//	if err != nil {
	//		return "", err
	//	}
	//	return "", nil
	//}
	//
	//// Upload a file (in gitter case send the upload URL because gitter has no native upload support)
	//if msg.Extra != nil {
	//	for _, rmsg := range helper.HandleExtra(&msg, b.General) {
	//		b.c.SendMessage(roomID, rmsg.Username+rmsg.Text)
	//	}
	//	if len(msg.Extra["file"]) > 0 {
	//		return b.handleUploadFile(&msg, roomID)
	//	}
	//}
	//
	//// Edit message
	//if msg.ID != "" {
	//	b.Log.Debugf("updating message with id %s", msg.ID)
	//	_, err := b.c.UpdateMessage(roomID, msg.ID, msg.Username+msg.Text)
	//	if err != nil {
	//		return "", err
	//	}
	//	return "", nil
	//}
	//
	//// Post normal message
	//resp, err := b.c.SendMessage(roomID, msg.Username+msg.Text)
	//if err != nil {
	//	return "", err
	//}
	//return resp.ID, nil

	b.Log.Debugf("=> Sending %#v", msg)

	err := b.conn.Send(text)

	// TODO return message id
	return "", err
}

// TODO do we want that? to allow login with QR code from a bridged channel? https://github.com/tulir/mautrix-whatsapp/blob/513eb18e2d59bada0dd515ee1abaaf38a3bfe3d5/commands.go#L76
//func (b *Bwhatsapp) Command(cmd string) string {
//	return ""
//}
