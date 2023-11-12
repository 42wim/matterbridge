package server

import (
	"bytes"
	"database/sql"
	"errors"
	"image"
	"image/color"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"

	eth_common "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/ipfs"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/protocol/identity/colorhash"
	"github.com/status-im/status-go/protocol/identity/ring"
	"github.com/status-im/status-go/services/wallet/bigint"
)

const (
	basePath                       = "/messages"
	imagesPath                     = basePath + "/images"
	audioPath                      = basePath + "/audio"
	ipfsPath                       = "/ipfs"
	discordAuthorsPath             = "/discord/authors"
	discordAttachmentsPath         = basePath + "/discord/attachments"
	LinkPreviewThumbnailPath       = "/link-preview/thumbnail"
	StatusLinkPreviewThumbnailPath = "/status-link-preview/thumbnail"
	communityTokenImagesPath       = "/communityTokenImages"

	walletBasePath              = "/wallet"
	walletCommunityImagesPath   = walletBasePath + "/communityImages"
	walletCollectionImagesPath  = walletBasePath + "/collectionImages"
	walletCollectibleImagesPath = walletBasePath + "/collectibleImages"

	// Handler routes for pairing
	accountImagesPath   = "/accountImages"
	accountInitialsPath = "/accountInitials"
	contactImagesPath   = "/contactImages"
	generateQRCode      = "/GenerateQRCode"
)

type HandlerPatternMap map[string]http.HandlerFunc

func handleRequestDBMissing(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Error("can't handle media request without appdb")
	}
}

func handleRequestDownloaderMissing(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Error("can't handle media request without ipfs downloader")
	}
}

type ImageParams struct {
	KeyUID                string
	PublicKey             string
	ImageName             string
	ImagePath             string
	FullName              string
	InitialsLength        int
	FontFile              string
	FontSize              float64
	Color                 color.Color
	BgSize                int
	BgColor               color.Color
	UppercaseRatio        float64
	Theme                 ring.Theme
	Ring                  bool
	RingWidth             float64
	IndicatorSize         float64
	IndicatorBorder       float64
	IndicatorCenterToEdge float64
	IndicatorColor        color.Color

	AuthorID     string
	URL          string
	MessageID    string
	AttachmentID string
	ImageID      string

	Hash     string
	Download bool
}

func ParseImageParams(logger *zap.Logger, params url.Values) ImageParams {
	parsed := ImageParams{}
	parsed.Color = color.Transparent
	parsed.BgColor = color.Transparent
	parsed.IndicatorColor = color.Transparent
	parsed.UppercaseRatio = 1.0

	keyUids := params["keyUid"]
	if len(keyUids) != 0 {
		parsed.KeyUID = keyUids[0]
	}

	pks := params["publicKey"]
	if len(pks) != 0 {
		parsed.PublicKey = pks[0]
	}

	imageNames := params["imageName"]
	if len(imageNames) != 0 {
		if filepath.IsAbs(imageNames[0]) {
			if _, err := os.Stat(imageNames[0]); err == nil {
				parsed.ImagePath = imageNames[0]
			} else if errors.Is(err, os.ErrNotExist) {
				logger.Error("ParseParams: image not exit", zap.String("imageName", imageNames[0]))
				return parsed
			} else {
				logger.Error("ParseParams: failed to read image", zap.String("imageName", imageNames[0]), zap.Error(err))
				return parsed
			}
		} else {
			parsed.ImageName = imageNames[0]
		}
	}

	names := params["name"]
	if len(names) != 0 {
		parsed.FullName = names[0]
	}

	parsed.InitialsLength = 2
	amountInitialsStr := params["length"]
	if len(amountInitialsStr) != 0 {
		amountInitials, err := strconv.Atoi(amountInitialsStr[0])
		if err != nil {
			logger.Error("ParseParams: invalid initials length")
			return parsed
		}
		parsed.InitialsLength = amountInitials
	}

	fontFiles := params["fontFile"]
	if len(fontFiles) != 0 {
		if _, err := os.Stat(fontFiles[0]); err == nil {
			parsed.FontFile = fontFiles[0]
		} else if errors.Is(err, os.ErrNotExist) {
			logger.Error("ParseParams: font file not exit", zap.String("FontFile", fontFiles[0]))
			return parsed
		} else {
			logger.Error("ParseParams: font file not exit", zap.String("FontFile", fontFiles[0]), zap.Error(err))
			return parsed
		}
	}

	fontSizeStr := params["fontSize"]
	if len(fontSizeStr) != 0 {
		fontSize, err := strconv.ParseFloat(fontSizeStr[0], 64)
		if err != nil {
			logger.Error("ParseParams: invalid fontSize", zap.String("FontSize", fontSizeStr[0]))
			return parsed
		}
		parsed.FontSize = fontSize
	}

	colors := params["color"]
	if len(colors) != 0 {
		textColor, err := images.ParseColor(colors[0])
		if err != nil {
			logger.Error("ParseParams: invalid color", zap.String("Color", colors[0]))
			return parsed
		}
		parsed.Color = textColor
	}

	sizeStrs := params["size"]
	if len(sizeStrs) != 0 {
		size, err := strconv.Atoi(sizeStrs[0])
		if err != nil {
			logger.Error("ParseParams: invalid size", zap.String("size", sizeStrs[0]))
			return parsed
		}
		parsed.BgSize = size
	}

	bgColors := params["bgColor"]
	if len(bgColors) != 0 {
		bgColor, err := images.ParseColor(bgColors[0])
		if err != nil {
			logger.Error("ParseParams: invalid bgColor", zap.String("BgColor", bgColors[0]))
			return parsed
		}
		parsed.BgColor = bgColor
	}

	uppercaseRatioStr := params["uppercaseRatio"]
	if len(uppercaseRatioStr) != 0 {
		uppercaseRatio, err := strconv.ParseFloat(uppercaseRatioStr[0], 64)
		if err != nil {
			logger.Error("ParseParams: invalid uppercaseRatio", zap.String("uppercaseRatio", uppercaseRatioStr[0]))
			return parsed
		}
		parsed.UppercaseRatio = uppercaseRatio
	}

	indicatorColors := params["indicatorColor"]
	if len(indicatorColors) != 0 {
		indicatorColor, err := images.ParseColor(indicatorColors[0])
		if err != nil {
			logger.Error("ParseParams: invalid indicatorColor", zap.String("IndicatorColor", indicatorColors[0]))
			return parsed
		}
		parsed.IndicatorColor = indicatorColor
	}

	indicatorSizeStrs := params["indicatorSize"]
	if len(indicatorSizeStrs) != 0 {
		indicatorSize, err := strconv.ParseFloat(indicatorSizeStrs[0], 64)
		if err != nil {
			logger.Error("ParseParams: invalid indicatorSize", zap.String("indicatorSize", indicatorSizeStrs[0]))
			indicatorSize = 0
		}
		parsed.IndicatorSize = indicatorSize
	}

	indicatorBorderStrs := params["indicatorBorder"]
	if len(indicatorBorderStrs) != 0 {
		indicatorBorder, err := strconv.ParseFloat(indicatorBorderStrs[0], 64)
		if err != nil {
			logger.Error("ParseParams: invalid indicatorBorder", zap.String("indicatorBorder", indicatorBorderStrs[0]))
			indicatorBorder = 0
		}
		parsed.IndicatorBorder = indicatorBorder
	}

	indicatorCenterToEdgeStrs := params["indicatorCenterToEdge"]
	if len(indicatorCenterToEdgeStrs) != 0 {
		indicatorCenterToEdge, err := strconv.ParseFloat(indicatorCenterToEdgeStrs[0], 64)
		if err != nil {
			logger.Error("ParseParams: invalid indicatorCenterToEdge", zap.String("indicatorCenterToEdge", indicatorCenterToEdgeStrs[0]))
			indicatorCenterToEdge = 0
		}
		parsed.IndicatorCenterToEdge = indicatorCenterToEdge
	}

	ringWidthStrs := params["ringWidth"]
	if len(ringWidthStrs) != 0 {
		ringWidth, err := strconv.ParseFloat(ringWidthStrs[0], 64)
		if err != nil {
			logger.Error("ParseParams: invalid indicatorSize", zap.String("ringWidth", ringWidthStrs[0]))
			ringWidth = 0
		}
		parsed.RingWidth = ringWidth
	}

	parsed.Theme = getTheme(params, logger)
	parsed.Ring = ringEnabled(params)

	messageIDs := params["message-id"]
	if len(messageIDs) != 0 {
		parsed.MessageID = messageIDs[0]
	}

	messageIDs = params["messageId"]
	if len(messageIDs) != 0 {
		parsed.MessageID = messageIDs[0]
	}

	authorIds := params["authorId"]
	if len(authorIds) != 0 {
		parsed.AuthorID = authorIds[0]
	}

	if attachmentIDs := params["attachmentId"]; len(attachmentIDs) != 0 {
		parsed.AttachmentID = attachmentIDs[0]
	}

	if imageIds := params["image-id"]; len(imageIds) != 0 {
		parsed.ImageID = imageIds[0]
	}

	urls := params["url"]
	if len(urls) != 0 {
		parsed.URL = urls[0]
	}

	hash := params["hash"]
	if len(hash) != 0 {
		parsed.Hash = hash[0]
	}

	_, download := params["download"]
	parsed.Download = download

	return parsed
}

func handleAccountImagesImpl(multiaccountsDB *multiaccounts.Database, logger *zap.Logger, w http.ResponseWriter, parsed ImageParams) {
	if parsed.KeyUID == "" {
		logger.Error("handleAccountImagesImpl: no keyUid")
		return
	}

	if parsed.ImageName == "" {
		logger.Error("handleAccountImagesImpl: no imageName")
		return
	}

	identityImage, err := multiaccountsDB.GetIdentityImage(parsed.KeyUID, parsed.ImageName)
	if err != nil {
		logger.Error("handleAccountImagesImpl: failed to load image.", zap.String("keyUid", parsed.KeyUID), zap.String("imageName", parsed.ImageName), zap.Error(err))
		return
	}

	if parsed.Ring && parsed.RingWidth == 0 {
		logger.Error("handleAccountImagesImpl: no ringWidth.")
		return
	}

	if parsed.BgSize == 0 {
		parsed.BgSize = identityImage.Width
	}

	payload, err := images.RoundCrop(identityImage.Payload)
	if err != nil {
		logger.Error("handleAccountImagesImpl: failed to crop image.", zap.String("keyUid", parsed.KeyUID), zap.String("imageName", parsed.ImageName), zap.Error(err))
		return
	}

	enlargeRatio := float64(identityImage.Width) / float64(parsed.BgSize)

	if parsed.Ring {
		account, err := multiaccountsDB.GetAccount(parsed.KeyUID)
		if err != nil {
			logger.Error("handleAccountImagesImpl: failed to GetAccount .", zap.String("keyUid", parsed.KeyUID), zap.Error(err))
			return
		}

		accColorHash := account.ColorHash

		if accColorHash == nil {
			if parsed.PublicKey == "" {
				logger.Error("handleAccountImagesImpl: no public key for color hash", zap.String("keyUid", parsed.KeyUID))
				return
			}

			accColorHash, err = colorhash.GenerateFor(parsed.PublicKey)
			if err != nil {
				logger.Error("handleAccountImagesImpl: could not generate color hash", zap.String("keyUid", parsed.KeyUID), zap.Error(err))
				return
			}
		}

		payload, err = ring.DrawRing(&ring.DrawRingParam{
			Theme: parsed.Theme, ColorHash: accColorHash, ImageBytes: payload, Height: identityImage.Height, Width: identityImage.Width, RingWidth: parsed.RingWidth * enlargeRatio,
		})
		if err != nil {
			logger.Error("handleAccountImagesImpl: failed to draw ring for account identity", zap.Error(err))
			return
		}
	}

	if parsed.IndicatorSize != 0 {
		// enlarge indicator size based on identity image size / desired size
		// or we get a bad quality identity image
		payload, err = images.AddStatusIndicatorToImage(payload, parsed.IndicatorColor, parsed.IndicatorSize*enlargeRatio, parsed.IndicatorBorder*enlargeRatio, parsed.IndicatorCenterToEdge*enlargeRatio)
		if err != nil {
			logger.Error("handleAccountImagesImpl: failed to draw status-indicator for initials", zap.Error(err))
			return
		}
	}

	if len(payload) == 0 {
		logger.Error("handleAccountImagesImpl: empty image")
		return
	}

	mime, err := images.GetProtobufImageMime(payload)
	if err != nil {
		logger.Error("failed to get mime", zap.Error(err))
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "no-store")

	_, err = w.Write(payload)
	if err != nil {
		logger.Error("handleAccountImagesImpl: failed to write image", zap.Error(err))
	}
}

func handleAccountImagesPlaceholder(logger *zap.Logger, w http.ResponseWriter, parsed ImageParams) {
	if parsed.ImagePath == "" {
		logger.Error("handleAccountImagesPlaceholder: no imagePath")
		return
	}

	payload, im, err := images.ImageToBytesAndImage(parsed.ImagePath)
	if err != nil {
		logger.Error("handleAccountImagesPlaceholder: failed to load image from disk", zap.String("imageName", parsed.ImagePath))
		return
	}
	width := im.Bounds().Dx()
	if parsed.BgSize == 0 {
		parsed.BgSize = width
	}

	payload, err = images.RoundCrop(payload)
	if err != nil {
		logger.Error("handleAccountImagesPlaceholder: failed to crop image.", zap.String("imageName", parsed.ImagePath), zap.Error(err))
		return
	}

	if parsed.IndicatorSize != 0 {
		enlargeIndicatorRatio := float64(width / parsed.BgSize)
		payload, err = images.AddStatusIndicatorToImage(payload, parsed.IndicatorColor, parsed.IndicatorSize*enlargeIndicatorRatio, parsed.IndicatorBorder*enlargeIndicatorRatio, parsed.IndicatorCenterToEdge)
		if err != nil {
			logger.Error("handleAccountImagesPlaceholder: failed to draw status-indicator for initials", zap.Error(err))
			return
		}
	}

	if len(payload) == 0 {
		logger.Error("handleAccountImagesPlaceholder: empty image")
		return
	}

	mime, err := images.GetProtobufImageMime(payload)
	if err != nil {
		logger.Error("failed to get mime", zap.Error(err))
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "no-store")

	_, err = w.Write(payload)
	if err != nil {
		logger.Error("handleAccountImagesPlaceholder: failed to write image", zap.Error(err))
	}
}

// handleAccountImages render multiaccounts custom profile image
func handleAccountImages(multiaccountsDB *multiaccounts.Database, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.KeyUID == "" {
			handleAccountImagesPlaceholder(logger, w, parsed)
		} else {
			handleAccountImagesImpl(multiaccountsDB, logger, w, parsed)
		}
	}
}

func handleAccountInitialsImpl(multiaccountsDB *multiaccounts.Database, logger *zap.Logger, w http.ResponseWriter, parsed ImageParams) {
	var name = parsed.FullName
	var accColorHash multiaccounts.ColorHash
	var account *multiaccounts.Account

	if parsed.Ring && parsed.RingWidth == 0 {
		logger.Error("handleAccountInitialsImpl: no ringWidth.")
		return
	}

	if parsed.KeyUID != "" {
		account, err := multiaccountsDB.GetAccount(parsed.KeyUID)

		if err != nil {
			logger.Error("handleAccountInitialsImpl: failed to get account.", zap.String("keyUid", parsed.KeyUID), zap.Error(err))
			return
		}
		name = account.Name
		accColorHash = account.ColorHash
	}

	initials := images.ExtractInitials(name, parsed.InitialsLength)

	payload, err := images.GenerateInitialsImage(initials, parsed.BgColor, parsed.Color, parsed.FontFile, parsed.BgSize, parsed.FontSize, parsed.UppercaseRatio)

	if err != nil {
		logger.Error("handleAccountInitialsImpl: failed to generate initials image.", zap.String("keyUid", parsed.KeyUID), zap.String("name", account.Name), zap.Error(err))
		return
	}

	if parsed.Ring {
		if accColorHash == nil {
			if parsed.PublicKey == "" {
				logger.Error("handleAccountInitialsImpl: no public key, can't draw ring", zap.String("keyUid", parsed.KeyUID), zap.Error(err))
				return
			}

			accColorHash, err = colorhash.GenerateFor(parsed.PublicKey)
			if err != nil {
				logger.Error("handleAccountInitialsImpl: failed to generate color hash from pubkey", zap.String("keyUid", parsed.KeyUID), zap.Error(err))
				return
			}
		}

		payload, err = ring.DrawRing(&ring.DrawRingParam{
			Theme: parsed.Theme, ColorHash: accColorHash, ImageBytes: payload, Height: parsed.BgSize, Width: parsed.BgSize, RingWidth: parsed.RingWidth,
		})

		if err != nil {
			logger.Error("failed to draw ring for account identity", zap.Error(err))
			return
		}
	}

	if parsed.IndicatorSize != 0 {
		payload, err = images.AddStatusIndicatorToImage(payload, parsed.IndicatorColor, parsed.IndicatorSize, parsed.IndicatorBorder, parsed.IndicatorCenterToEdge)
		if err != nil {
			logger.Error("failed to draw status-indicator for initials", zap.Error(err))
			return
		}
	}

	if len(payload) == 0 {
		logger.Error("handleAccountInitialsImpl: empty image", zap.String("keyUid", parsed.KeyUID), zap.Error(err))
		return
	}
	mime, err := images.GetProtobufImageMime(payload)
	if err != nil {
		logger.Error("failed to get mime", zap.Error(err))
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "no-store")

	_, err = w.Write(payload)
	if err != nil {
		logger.Error("failed to write image", zap.Error(err))
	}
}

func handleAccountInitialsPlaceholder(logger *zap.Logger, w http.ResponseWriter, parsed ImageParams) {
	if parsed.FullName == "" {
		logger.Error("handleAccountInitialsPlaceholder: no full name")
		return
	}

	initials := images.ExtractInitials(parsed.FullName, parsed.InitialsLength)

	payload, err := images.GenerateInitialsImage(initials, parsed.BgColor, parsed.Color, parsed.FontFile, parsed.BgSize, parsed.FontSize, parsed.UppercaseRatio)

	if err != nil {
		logger.Error("handleAccountInitialsPlaceholder: failed to generate initials image.", zap.String("keyUid", parsed.KeyUID), zap.String("name", parsed.FullName), zap.Error(err))
		return
	}

	if parsed.IndicatorSize != 0 {
		payload, err = images.AddStatusIndicatorToImage(payload, parsed.IndicatorColor, parsed.IndicatorSize, parsed.IndicatorBorder, parsed.IndicatorCenterToEdge)
		if err != nil {
			logger.Error("failed to draw status-indicator for initials", zap.Error(err))
			return
		}
	}

	if len(payload) == 0 {
		logger.Error("handleAccountInitialsPlaceholder: empty image", zap.String("keyUid", parsed.KeyUID), zap.Error(err))
		return
	}
	mime, err := images.GetProtobufImageMime(payload)
	if err != nil {
		logger.Error("failed to get mime", zap.Error(err))
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "no-store")

	_, err = w.Write(payload)
	if err != nil {
		logger.Error("failed to write image", zap.Error(err))
	}
}

// handleAccountInitials render multiaccounts/contacts initials avatar image
func handleAccountInitials(multiaccountsDB *multiaccounts.Database, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.FontFile == "" {
			logger.Error("handleAccountInitials: no fontFile")
			return
		}
		if parsed.FontSize == 0 {
			logger.Error("handleAccountInitials: no fontSize")
			return
		}
		if parsed.Color == color.Transparent {
			logger.Error("handleAccountInitials: no color")
			return
		}
		if parsed.BgSize == 0 {
			logger.Error("handleAccountInitials: no size")
			return
		}
		if parsed.BgColor == color.Transparent {
			logger.Error("handleAccountInitials: no bgColor")
			return
		}

		if parsed.KeyUID == "" && parsed.PublicKey == "" {
			handleAccountInitialsPlaceholder(logger, w, parsed)
		} else {
			handleAccountInitialsImpl(multiaccountsDB, logger, w, parsed)
		}
	}
}

// handleContactImages render contacts custom profile image
func handleContactImages(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.PublicKey == "" {
			logger.Error("no publicKey")
			return
		}

		if parsed.ImageName == "" {
			logger.Error("no imageName")
			return
		}

		if parsed.Ring && parsed.RingWidth == 0 {
			logger.Error("handleAccountImagesImpl: no ringWidth.")
			return
		}

		var payload []byte
		err := db.QueryRow(`SELECT payload FROM chat_identity_contacts WHERE contact_id = ? and image_type = ?`, parsed.PublicKey, parsed.ImageName).Scan(&payload)
		if err != nil {
			logger.Error("failed to load image.", zap.String("contact id", parsed.PublicKey), zap.String("image type", parsed.ImageName), zap.Error(err))
			return
		}

		img, _, err := image.Decode(bytes.NewReader(payload))
		if err != nil {
			logger.Error("failed to decode config.", zap.String("contact id", parsed.PublicKey), zap.String("image type", parsed.ImageName), zap.Error(err))
			return
		}
		width := img.Bounds().Dx()

		if parsed.BgSize == 0 {
			parsed.BgSize = width
		}

		payload, err = images.RoundCrop(payload)
		if err != nil {
			logger.Error("handleContactImages: failed to crop image.", zap.Error(err))
			return
		}

		enlargeRatio := float64(width) / float64(parsed.BgSize)

		if parsed.Ring {
			colorHash, err := colorhash.GenerateFor(parsed.PublicKey)
			if err != nil {
				logger.Error("could not generate color hash")
				return
			}

			payload, err = ring.DrawRing(&ring.DrawRingParam{
				Theme: parsed.Theme, ColorHash: colorHash, ImageBytes: payload, Height: width, Width: width, RingWidth: parsed.RingWidth * enlargeRatio,
			})

			if err != nil {
				logger.Error("failed to draw ring for contact image.", zap.Error(err))
				return
			}
		}

		if parsed.IndicatorSize != 0 {
			payload, err = images.AddStatusIndicatorToImage(payload, parsed.IndicatorColor, parsed.IndicatorSize*enlargeRatio, parsed.IndicatorBorder*enlargeRatio, parsed.IndicatorCenterToEdge*enlargeRatio)
			if err != nil {
				logger.Error("handleAccountImagesImpl: failed to draw status-indicator for initials", zap.Error(err))
				return
			}
		}

		if len(payload) == 0 {
			logger.Error("empty image")
			return
		}
		mime, err := images.GetProtobufImageMime(payload)
		if err != nil {
			logger.Error("failed to get mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(payload)
		if err != nil {
			logger.Error("failed to write image", zap.Error(err))
		}
	}
}

func ringEnabled(params url.Values) bool {
	addRings, ok := params["addRing"]
	return ok && len(addRings) == 1 && addRings[0] == "1"
}

func getTheme(params url.Values, logger *zap.Logger) ring.Theme {
	theme := ring.LightTheme // default
	themes, ok := params["theme"]
	if ok && len(themes) > 0 {
		t, err := strconv.Atoi(themes[0])
		if err != nil {
			logger.Error("invalid param[theme], value: " + themes[0])
		} else {
			theme = ring.Theme(t)
		}
	}
	return theme
}

func handleDiscordAuthorAvatar(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.AuthorID == "" {
			logger.Error("no authorIDs")
			return
		}

		var image []byte
		err := db.QueryRow(`SELECT avatar_image_payload FROM discord_message_authors WHERE id = ?`, parsed.AuthorID).Scan(&image)
		if err != nil {
			logger.Error("failed to find image", zap.Error(err))
			return
		}
		if len(image) == 0 {
			logger.Error("empty image")
			return
		}
		mime, err := images.GetProtobufImageMime(image)
		if err != nil {
			logger.Error("failed to get mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(image)
		if err != nil {
			logger.Error("failed to write image", zap.Error(err))
		}
	}
}

func handleDiscordAttachment(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.MessageID == "" {
			logger.Error("no messageID")
			return
		}
		if parsed.AttachmentID == "" {
			logger.Error("no attachmentID")
			return
		}

		var image []byte
		err := db.QueryRow(`SELECT payload FROM discord_message_attachments WHERE discord_message_id = ? AND id = ?`, parsed.MessageID, parsed.AttachmentID).Scan(&image)
		if err != nil {
			logger.Error("failed to find image", zap.Error(err))
			return
		}
		if len(image) == 0 {
			logger.Error("empty image")
			return
		}
		mime, err := images.GetProtobufImageMime(image)
		if err != nil {
			logger.Error("failed to get mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(image)
		if err != nil {
			logger.Error("failed to write image", zap.Error(err))
		}
	}
}

func handleImage(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.MessageID == "" {
			logger.Error("no messageID")
			return
		}

		var image []byte
		err := db.QueryRow(`SELECT image_payload FROM user_messages WHERE id = ?`, parsed.MessageID).Scan(&image)
		if err != nil {
			logger.Error("failed to find image", zap.Error(err))
			return
		}
		if len(image) == 0 {
			logger.Error("empty image")
			return
		}
		mime, err := images.GetProtobufImageMime(image)
		if err != nil {
			logger.Error("failed to get mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(image)
		if err != nil {
			logger.Error("failed to write image", zap.Error(err))
		}
	}
}

func handleAudio(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.MessageID == "" {
			logger.Error("no messageID")
			return
		}

		var audio []byte
		err := db.QueryRow(`SELECT audio_payload FROM user_messages WHERE id = ?`, parsed.MessageID).Scan(&audio)
		if err != nil {
			logger.Error("failed to find image", zap.Error(err))
			return
		}
		if len(audio) == 0 {
			logger.Error("empty audio")
			return
		}

		w.Header().Set("Content-Type", "audio/aac")
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(audio)
		if err != nil {
			logger.Error("failed to write audio", zap.Error(err))
		}
	}
}

func handleIPFS(downloader *ipfs.Downloader, logger *zap.Logger) http.HandlerFunc {
	if downloader == nil {
		return handleRequestDownloaderMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		parsed := ParseImageParams(logger, params)

		if parsed.Hash == "" {
			logger.Error("no hash")
			return
		}

		content, err := downloader.Get(parsed.Hash, parsed.Download)
		if err != nil {
			logger.Error("could not download hash", zap.Error(err))
			return
		}

		w.Header().Set("Cache-Control", "max-age:290304000, public")
		w.Header().Set("Expires", time.Now().AddDate(60, 0, 0).Format(http.TimeFormat))

		_, err = w.Write(content)
		if err != nil {
			logger.Error("failed to write ipfs resource", zap.Error(err))
		}
	}
}

func handleQRCodeGeneration(multiaccountsDB *multiaccounts.Database, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		payload := generateQRBytes(params, logger, multiaccountsDB)
		mime, err := images.GetProtobufImageMime(payload)

		if err != nil {
			logger.Error("could not generate image from payload", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(payload)

		if err != nil {
			logger.Error("failed to write image", zap.Error(err))
		}
	}
}

func handleCommunityTokenImages(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		if len(params["communityID"]) == 0 {
			logger.Error("no communityID")
			return
		}
		if len(params["chainID"]) == 0 {
			logger.Error("no chainID")
			return
		}
		if len(params["symbol"]) == 0 {
			logger.Error("no symbol")
			return
		}

		chainID, err := strconv.ParseUint(params["chainID"][0], 10, 64)
		if err != nil {
			logger.Error("invalid chainID in community token image", zap.Error(err))
			return
		}

		var base64Image string
		err = db.QueryRow("SELECT image_base64 FROM community_tokens WHERE community_id = ? AND chain_id = ? AND symbol = ?", params["communityID"][0], chainID, params["symbol"][0]).Scan(&base64Image)
		if err != nil {
			logger.Error("failed to find community token image", zap.Error(err))
			return
		}
		if len(base64Image) == 0 {
			logger.Error("empty community token image")
			return
		}
		imagePayload, err := images.GetPayloadFromURI(base64Image)
		if err != nil {
			logger.Error("failed to get community token image payload", zap.Error(err))
			return
		}
		mime, err := images.GetProtobufImageMime(imagePayload)
		if err != nil {
			logger.Error("failed to get community token image mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(imagePayload)
		if err != nil {
			logger.Error("failed to write community token image", zap.Error(err))
		}
	}
}

func handleWalletCommunityImages(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		if len(params["communityID"]) == 0 {
			logger.Error("no communityID")
			return
		}

		var image []byte
		err := db.QueryRow(`SELECT image_payload FROM community_data_cache WHERE id = ?`, params["communityID"][0]).Scan(&image)
		if err != nil {
			logger.Error("failed to find wallet community image", zap.Error(err))
			return
		}
		if len(image) == 0 {
			logger.Error("empty wallet community image")
			return
		}
		mime, err := images.GetProtobufImageMime(image)
		if err != nil {
			logger.Error("failed to get wallet community image mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(image)
		if err != nil {
			logger.Error("failed to write wallet community image", zap.Error(err))
		}
	}
}

func handleWalletCollectionImages(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		if len(params["chainID"]) == 0 {
			logger.Error("no chainID")
			return
		}

		if len(params["contractAddress"]) == 0 {
			logger.Error("no contractAddress")
			return
		}

		chainID, err := strconv.ParseUint(params["chainID"][0], 10, 64)
		if err != nil {
			logger.Error("invalid chainID in wallet collectible image", zap.Error(err))
			return
		}
		contractAddress := eth_common.HexToAddress(params["contractAddress"][0])
		if len(contractAddress) == 0 {
			logger.Error("invalid contractAddress in wallet collectible image", zap.Error(err))
			return
		}

		var image []byte
		err = db.QueryRow(`SELECT image_payload FROM collection_data_cache WHERE chain_id = ? AND contract_address = ?`,
			chainID,
			contractAddress).Scan(&image)
		if err != nil {
			logger.Error("failed to find wallet collection image", zap.Error(err))
			return
		}
		if len(image) == 0 {
			logger.Error("empty wallet collection image")
			return
		}
		mime, err := images.GetProtobufImageMime(image)
		if err != nil {
			logger.Error("failed to get wallet collection image mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(image)
		if err != nil {
			logger.Error("failed to write wallet collection image", zap.Error(err))
		}
	}
}

func handleWalletCollectibleImages(db *sql.DB, logger *zap.Logger) http.HandlerFunc {
	if db == nil {
		return handleRequestDBMissing(logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		if len(params["chainID"]) == 0 {
			logger.Error("no chainID")
			return
		}

		if len(params["contractAddress"]) == 0 {
			logger.Error("no contractAddress")
			return
		}

		if len(params["tokenID"]) == 0 {
			logger.Error("no tokenID")
			return
		}

		chainID, err := strconv.ParseUint(params["chainID"][0], 10, 64)
		if err != nil {
			logger.Error("invalid chainID in wallet collectible image", zap.Error(err))
			return
		}
		contractAddress := eth_common.HexToAddress(params["contractAddress"][0])
		if len(contractAddress) == 0 {
			logger.Error("invalid contractAddress in wallet collectible image", zap.Error(err))
			return
		}
		tokenID, ok := big.NewInt(0).SetString(params["tokenID"][0], 10)
		if !ok {
			logger.Error("invalid tokenID in wallet collectible image", zap.Error(err))
			return
		}

		var image []byte
		err = db.QueryRow(`SELECT image_payload FROM collectible_data_cache WHERE chain_id = ? AND contract_address = ? AND token_id = ?`,
			chainID,
			contractAddress,
			(*bigint.SQLBigIntBytes)(tokenID)).Scan(&image)
		if err != nil {
			logger.Error("failed to find wallet collectible image", zap.Error(err))
			return
		}
		if len(image) == 0 {
			logger.Error("empty image")
			return
		}
		mime, err := images.GetProtobufImageMime(image)
		if err != nil {
			logger.Error("failed to get wallet collectible image mime", zap.Error(err))
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")

		_, err = w.Write(image)
		if err != nil {
			logger.Error("failed to write wallet collectible image", zap.Error(err))
		}
	}
}
