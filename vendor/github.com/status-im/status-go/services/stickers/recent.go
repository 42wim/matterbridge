package stickers

import (
	"encoding/json"

	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/services/wallet/bigint"
)

const maxNumberRecentStickers = 24

func (api *API) recentStickers() ([]Sticker, error) {
	installedStickersPacksJSON, err := api.accountsDB.GetInstalledStickerPacks()

	if err != nil || installedStickersPacksJSON == nil {
		return []Sticker{}, nil
	}

	recentStickersJSON, err := api.accountsDB.GetRecentStickers()

	if err != nil || recentStickersJSON == nil {
		return []Sticker{}, nil
	}

	recentStickersList := make([]Sticker, 0)
	if err := json.Unmarshal(*recentStickersJSON, &recentStickersList); err != nil {
		return []Sticker{}, err
	}

	var installedStickersPacks map[string]StickerPack
	if err := json.Unmarshal(*installedStickersPacksJSON, &installedStickersPacks); err != nil {
		return []Sticker{}, err
	}

	recentStickersListInExistingPacks := make([]Sticker, 0)
	existingPackIDs := make(map[string]bool)

	for k := range installedStickersPacks {
		existingPackIDs[k] = true
	}

	for _, s := range recentStickersList {
		packIDStr := s.PackID.String()
		if _, exists := existingPackIDs[packIDStr]; exists {
			recentStickersListInExistingPacks = append(recentStickersListInExistingPacks, s)
		}
	}

	return recentStickersListInExistingPacks, nil
}

func (api *API) ClearRecent() error {
	var recentStickersList []Sticker
	return api.accountsDB.SaveSettingField(settings.StickersRecentStickers, recentStickersList)
}

func (api *API) Recent() ([]Sticker, error) {
	recentStickersList, err := api.recentStickers()
	if err != nil {
		return nil, err
	}

	for i, sticker := range recentStickersList {
		sticker.URL = api.hashToURL(sticker.Hash)
		recentStickersList[i] = sticker
	}

	return recentStickersList, nil
}

func (api *API) AddRecent(packID *bigint.BigInt, hash string) error {
	sticker := Sticker{
		PackID: packID,
		Hash:   hash,
	}

	recentStickersList, err := api.recentStickers()
	if err != nil {
		return err
	}

	// Remove duplicated
	idx := -1
	for i, currSticker := range recentStickersList {
		if currSticker.PackID.Cmp(sticker.PackID.Int) == 0 && currSticker.Hash == sticker.Hash {
			idx = i
		}
	}
	if idx > -1 {
		recentStickersList = append(recentStickersList[:idx], recentStickersList[idx+1:]...)
	}

	sticker.URL = ""

	if len(recentStickersList) >= maxNumberRecentStickers {
		recentStickersList = append([]Sticker{sticker}, recentStickersList[:maxNumberRecentStickers-1]...)
	} else {
		recentStickersList = append([]Sticker{sticker}, recentStickersList...)
	}

	return api.accountsDB.SaveSettingField(settings.StickersRecentStickers, recentStickersList)
}
