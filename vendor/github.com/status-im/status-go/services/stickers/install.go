package stickers

import (
	"encoding/json"
	"errors"

	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/services/wallet/bigint"
)

func (api *API) Install(chainID uint64, packID *bigint.BigInt) error {
	installedPacks, err := api.installedStickerPacks()
	if err != nil {
		return err
	}

	if _, exists := installedPacks[uint(packID.Uint64())]; exists {
		return errors.New("sticker pack is already installed")
	}

	// TODO: this does not validate if the pack is purchased. Should it?

	stickerType, err := api.contractMaker.NewStickerType(chainID)
	if err != nil {
		return err
	}

	stickerPack, err := api.fetchPackData(stickerType, packID.Int, false)
	if err != nil {
		return err
	}

	installedPacks[uint(packID.Uint64())] = *stickerPack

	err = api.accountsDB.SaveSettingField(settings.StickersPacksInstalled, installedPacks)
	if err != nil {
		return err
	}

	return nil
}

func (api *API) installedStickerPacks() (StickerPackCollection, error) {
	stickerPacks := make(StickerPackCollection)

	installedStickersJSON, err := api.accountsDB.GetInstalledStickerPacks()
	if err != nil {
		return nil, err
	}

	if installedStickersJSON == nil {
		return stickerPacks, nil
	}

	err = json.Unmarshal(*installedStickersJSON, &stickerPacks)
	if err != nil {
		return nil, err
	}

	return stickerPacks, nil
}

func (api *API) Installed() (StickerPackCollection, error) {
	stickerPacks, err := api.installedStickerPacks()
	if err != nil {
		return nil, err
	}

	for packID, stickerPack := range stickerPacks {
		stickerPack.Status = statusInstalled
		stickerPack.Preview = api.hashToURL(stickerPack.Preview)
		stickerPack.Thumbnail = api.hashToURL(stickerPack.Thumbnail)
		for i, sticker := range stickerPack.Stickers {
			sticker.URL = api.hashToURL(sticker.Hash)
			if err != nil {
				return nil, err
			}
			stickerPack.Stickers[i] = sticker
		}
		stickerPacks[packID] = stickerPack
	}

	return stickerPacks, nil
}

func (api *API) Uninstall(packID *bigint.BigInt) error {
	installedPacks, err := api.installedStickerPacks()
	if err != nil {
		return err
	}

	if _, exists := installedPacks[uint(packID.Uint64())]; !exists {
		return errors.New("sticker pack is not installed")
	}

	delete(installedPacks, uint(packID.Uint64()))

	err = api.accountsDB.SaveSettingField(settings.StickersPacksInstalled, installedPacks)
	if err != nil {
		return err
	}

	// Removing uninstalled pack from recent stickers

	recentStickers, err := api.recentStickers()
	if err != nil {
		return err
	}

	idx := -1
	for i, r := range recentStickers {
		if r.PackID.Cmp(packID.Int) == 0 {
			idx = i
			break
		}
	}

	if idx > -1 {
		var newRecentStickers []Sticker
		newRecentStickers = append(newRecentStickers, recentStickers[:idx]...)
		if idx != len(recentStickers)-1 {
			newRecentStickers = append(newRecentStickers, recentStickers[idx+1:]...)
		}
		return api.accountsDB.SaveSettingField(settings.StickersRecentStickers, newRecentStickers)
	}

	return nil
}
