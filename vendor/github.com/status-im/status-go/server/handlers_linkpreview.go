package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

func getThumbnailPayload(db *sql.DB, logger *zap.Logger, msgID string, thumbnailURL string) ([]byte, error) {
	var payload []byte

	var result []byte
	err := db.QueryRow(`SELECT unfurled_links FROM user_messages WHERE id = ?`, msgID).Scan(&result)
	if err != nil {
		return payload, fmt.Errorf("could not find message with message-id '%s': %w", msgID, err)
	}

	var links []*protobuf.UnfurledLink
	err = json.Unmarshal(result, &links)
	if err != nil {
		return payload, fmt.Errorf("failed to unmarshal protobuf.UrlPreview: %w", err)
	}

	for _, p := range links {
		if p.Url == thumbnailURL {
			payload = p.ThumbnailPayload
			break
		}
	}

	return payload, nil
}

func handleLinkPreviewThumbnail(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.MessageID == "" {
			http.Error(w, "missing query parameter 'message-id'", http.StatusBadRequest)
			return
		}

		if parsed.URL == "" {
			http.Error(w, "missing query parameter 'url'", http.StatusBadRequest)
			return
		}

		thumbnail, err := getThumbnailPayload(db, logger, parsed.MessageID, parsed.URL)
		if err != nil {
			logger.Error("failed to get thumbnail", zap.String("msgID", parsed.MessageID))
			http.Error(w, "failed to get thumbnail", http.StatusInternalServerError)
			return
		}

		mimeType, err := images.GetMimeType(thumbnail)
		if err != nil {
			http.Error(w, "mime type not supported", http.StatusNotImplemented)
			return
		}

		w.Header().Set("Content-Type", "image/"+mimeType)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(thumbnail)
		if err != nil {
			logger.Error("failed to write response", zap.Error(err))
		}
	}
}

func getStatusLinkPreviewImage(p *protobuf.UnfurledStatusLink, imageID common.MediaServerImageID) ([]byte, error) {

	switch imageID {
	case common.MediaServerContactIcon:
		contact := p.GetContact()
		if contact == nil {
			return nil, fmt.Errorf("this is not a contact link")
		}
		if contact.Icon == nil {
			return nil, fmt.Errorf("contact icon is empty")
		}
		return contact.Icon.Payload, nil

	case common.MediaServerCommunityIcon:
		community := p.GetCommunity()
		if community == nil {
			return nil, fmt.Errorf("this is not a community link")
		}
		if community.Icon == nil {
			return nil, fmt.Errorf("community icon is empty")
		}
		return community.Icon.Payload, nil

	case common.MediaServerCommunityBanner:
		community := p.GetCommunity()
		if community == nil {
			return nil, fmt.Errorf("this is not a community link")
		}
		if community.Banner == nil {
			return nil, fmt.Errorf("community banner is empty")
		}
		return community.Banner.Payload, nil

	case common.MediaServerChannelCommunityIcon:
		channel := p.GetChannel()
		if channel == nil {
			return nil, fmt.Errorf("this is not a community channel link")
		}
		if channel.Community == nil {
			return nil, fmt.Errorf("channel community is empty")
		}
		if channel.Community.Icon == nil {
			return nil, fmt.Errorf("channel community icon is empty")
		}
		return channel.Community.Icon.Payload, nil

	case common.MediaServerChannelCommunityBanner:
		channel := p.GetChannel()
		if channel == nil {
			return nil, fmt.Errorf("this is not a community channel link")
		}
		if channel.Community == nil {
			return nil, fmt.Errorf("channel community is empty")
		}
		if channel.Community.Banner == nil {
			return nil, fmt.Errorf("channel community banner is empty")
		}
		return channel.Community.Banner.Payload, nil
	}

	return nil, fmt.Errorf("value not supported")
}

func getStatusLinkPreviewThumbnail(db *sql.DB, messageID string, URL string, imageID common.MediaServerImageID) ([]byte, int, error) {
	var messageLinks []byte
	err := db.QueryRow(`SELECT unfurled_status_links FROM user_messages WHERE id = ?`, messageID).Scan(&messageLinks)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("could not find message with message-id '%s': %w", messageID, err)
	}

	var links protobuf.UnfurledStatusLinks
	err = proto.Unmarshal(messageLinks, &links)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to unmarshal protobuf.UrlPreview: %w", err)
	}

	for _, p := range links.UnfurledStatusLinks {
		if p.Url == URL {
			thumbnailPayload, err := getStatusLinkPreviewImage(p, imageID)
			if err != nil {
				return nil, http.StatusBadRequest, fmt.Errorf("invalid query parameter 'image-id' value: %w", err)
			}
			return thumbnailPayload, http.StatusOK, nil
		}
	}

	return nil, http.StatusBadRequest, fmt.Errorf("no link preview found for given url")
}

func handleStatusLinkPreviewThumbnail(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.MessageID == "" {
			http.Error(w, "missing query parameter 'message-id'", http.StatusBadRequest)
			return
		}

		if parsed.URL == "" {
			http.Error(w, "missing query parameter 'url'", http.StatusBadRequest)
			return
		}

		if parsed.ImageID == "" {
			http.Error(w, "missing query parameter 'image-id'", http.StatusBadRequest)
			return
		}

		thumbnail, httpsStatusCode, err := getStatusLinkPreviewThumbnail(db, parsed.MessageID, parsed.URL, common.MediaServerImageID(parsed.ImageID))
		if err != nil {
			http.Error(w, err.Error(), httpsStatusCode)
			return
		}

		mimeType, err := images.GetMimeType(thumbnail)
		if err != nil {
			http.Error(w, "mime type not supported", http.StatusNotImplemented)
			return
		}

		w.Header().Set("Content-Type", "image/"+mimeType)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(thumbnail)
		if err != nil {
			logger.Error("failed to write response", zap.Error(err))
		}
	}
}
