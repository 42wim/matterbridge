package pushnotificationserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

const defaultNewMessageNotificationText = "You have a new message"
const defaultMentionNotificationText = "Someone mentioned you"
const defaultRequestToJoinCommunityNotificationText = "Someone requested to join a community you are an admin of"

type GoRushRequestData struct {
	EncryptedMessage string `json:"encryptedMessage"`
	ChatID           string `json:"chatId"`
	PublicKey        string `json:"publicKey"`
}

type GoRushRequestNotification struct {
	Tokens   []string           `json:"tokens"`
	Platform uint               `json:"platform"`
	Message  string             `json:"message"`
	Topic    string             `json:"topic"`
	Data     *GoRushRequestData `json:"data"`
}

type GoRushRequest struct {
	Notifications []*GoRushRequestNotification `json:"notifications"`
}

type RequestAndRegistration struct {
	Request      *protobuf.PushNotification
	Registration *protobuf.PushNotificationRegistration
}

func tokenTypeToGoRushPlatform(tokenType protobuf.PushNotificationRegistration_TokenType) uint {
	switch tokenType {
	case protobuf.PushNotificationRegistration_APN_TOKEN:
		return 1
	case protobuf.PushNotificationRegistration_FIREBASE_TOKEN:
		return 2
	}
	return 0
}

func PushNotificationRegistrationToGoRushRequest(requestAndRegistrations []*RequestAndRegistration) *GoRushRequest {
	goRushRequests := &GoRushRequest{}
	for _, requestAndRegistration := range requestAndRegistrations {
		request := requestAndRegistration.Request
		registration := requestAndRegistration.Registration
		var text string
		if request.Type == protobuf.PushNotification_MESSAGE {
			text = defaultNewMessageNotificationText
		} else if request.Type == protobuf.PushNotification_REQUEST_TO_JOIN_COMMUNITY {
			text = defaultRequestToJoinCommunityNotificationText
		} else {
			text = defaultMentionNotificationText
		}
		goRushRequests.Notifications = append(goRushRequests.Notifications,
			&GoRushRequestNotification{
				Tokens:   []string{registration.DeviceToken},
				Platform: tokenTypeToGoRushPlatform(registration.TokenType),
				Message:  text,
				Topic:    registration.ApnTopic,
				Data: &GoRushRequestData{
					EncryptedMessage: types.EncodeHex(request.Message),
					ChatID:           types.EncodeHex(request.ChatId),
					PublicKey:        types.EncodeHex(request.PublicKey),
				},
			})
	}
	return goRushRequests
}

func sendGoRushNotification(request *GoRushRequest, url string, logger *zap.Logger) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	response, err := http.Post(url+"/api/push", "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	logger.Info("Sent gorush request", zap.String("response", string(body)))

	return nil
}
