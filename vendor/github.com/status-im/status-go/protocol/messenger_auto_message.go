package protocol

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/signal"
)

// autoMessageInterval is how often we should send a message
const autoMessageInterval = 120 * time.Second

const autoMessageChatID = "status-bot"

func (m *Messenger) AutoMessageEnabled() (bool, error) {
	return m.settings.AutoMessageEnabled()
}

func (m *Messenger) startAutoMessageLoop() error {
	enabled, err := m.AutoMessageEnabled()
	if err != nil {
		m.logger.Error("[auto message] failed to start auto message loop", zap.Error(err))
		return err
	}

	if !enabled {
		return nil
	}

	m.logger.Info("[auto message] starting auto message loop")
	ticker := time.NewTicker(autoMessageInterval)
	count := 0
	go func() {
		for {
			select {
			case <-ticker.C:
				count++
				timestamp := time.Now().Format(time.RFC3339)

				msg := common.NewMessage()
				msg.Text = fmt.Sprintf("%d\n%s", count, timestamp)
				msg.ChatId = autoMessageChatID
				msg.LocalChatID = autoMessageChatID
				msg.ContentType = protobuf.ChatMessage_TEXT_PLAIN
				resp, err := m.SendChatMessage(context.Background(), msg)
				if err != nil {
					m.logger.Error("[auto message] failed to send message", zap.Error(err))
					continue
				}
				signal.SendNewMessages(resp)

				err = m.UpdateMessageOutgoingStatus(msg.ID, common.OutgoingStatusDelivered)
				if err != nil {
					m.logger.Error("[auto message] failed to mark message as delivered", zap.Error(err))
					continue
				}

				//send signal to client that message status updated
				if m.config.messengerSignalsHandler != nil {
					m.config.messengerSignalsHandler.MessageDelivered(autoMessageChatID, msg.ID)
				}
			case <-m.quit:
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}
