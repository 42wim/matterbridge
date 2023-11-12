package signal

import (
	"github.com/status-im/status-go/protocol/discord"
)

const (

	// EventDiscordCategoriesAndChannelsExtracted triggered when categories and
	// channels for exported discord files have been successfully extracted
	EventDiscordCategoriesAndChannelsExtracted = "community.discordCategoriesAndChannelsExtracted"

	// EventDiscordCommunityImportProgress is triggered during the import
	// of a discord community as it progresses
	EventDiscordCommunityImportProgress = "community.discordCommunityImportProgress"

	// EventDiscordCommunityImportFinished triggered when importing
	// the discord community into status was successful
	EventDiscordCommunityImportFinished = "community.discordCommunityImportFinished"

	// EventDiscordCommunityImportCancelled triggered when importing
	// the discord community was cancelled
	EventDiscordCommunityImportCancelled = "community.discordCommunityImportCancelled"

	// EventDiscordCommunityImportCleanedUp triggered when the community has been cleaned up (deleted)
	EventDiscordCommunityImportCleanedUp = "community.discordCommunityImportCleanedUp"

	// EventDiscordChannelImportProgress is triggered during the import
	// of a discord community channel as it progresses
	EventDiscordChannelImportProgress = "community.discordChannelImportProgress"

	// EventDiscordChannelImportFinished triggered when importing
	// the discord community channel into status was successful
	EventDiscordChannelImportFinished = "community.discordChannelImportFinished"

	// EventDiscordChannelImportCancelled triggered when importing
	// the discord community channel was cancelled
	EventDiscordChannelImportCancelled = "community.discordChannelImportCancelled"
)

type DiscordCategoriesAndChannelsExtractedSignal struct {
	Categories             []*discord.Category             `json:"discordCategories"`
	Channels               []*discord.Channel              `json:"discordChannels"`
	OldestMessageTimestamp int64                           `json:"oldestMessageTimestamp"`
	Errors                 map[string]*discord.ImportError `json:"errors"`
}

type DiscordCommunityImportProgressSignal struct {
	ImportProgress *discord.ImportProgress `json:"importProgress"`
}

type DiscordCommunityImportFinishedSignal struct {
	CommunityID string `json:"communityId"`
}

type DiscordCommunityImportCancelledSignal struct {
	CommunityID string `json:"communityId"`
}

type DiscordCommunityImportCleanedUpSignal struct {
	CommunityID string `json:"communityId"`
}

type DiscordChannelImportProgressSignal struct {
	ImportProgress *discord.ImportProgress `json:"importProgress"`
}

type DiscordChannelImportFinishedSignal struct {
	CommunityID string `json:"communityId"`
	ChannelID   string `json:"channelId"`
}

type DiscordChannelImportCancelledSignal struct {
	ChannelID string `json:"channelId"`
}

func SendDiscordCategoriesAndChannelsExtracted(categories []*discord.Category, channels []*discord.Channel, oldestMessageTimestamp int64, errors map[string]*discord.ImportError) {
	send(EventDiscordCategoriesAndChannelsExtracted, DiscordCategoriesAndChannelsExtractedSignal{
		Categories:             categories,
		Channels:               channels,
		OldestMessageTimestamp: oldestMessageTimestamp,
		Errors:                 errors,
	})
}

func SendDiscordCommunityImportProgress(importProgress *discord.ImportProgress) {
	send(EventDiscordCommunityImportProgress, DiscordCommunityImportProgressSignal{
		ImportProgress: importProgress,
	})
}

func SendDiscordChannelImportProgress(importProgress *discord.ImportProgress) {
	send(EventDiscordChannelImportProgress, DiscordChannelImportProgressSignal{
		ImportProgress: importProgress,
	})
}

func SendDiscordCommunityImportFinished(communityID string) {
	send(EventDiscordCommunityImportFinished, DiscordCommunityImportFinishedSignal{
		CommunityID: communityID,
	})
}

func SendDiscordChannelImportFinished(communityID string, channelID string) {
	send(EventDiscordChannelImportFinished, DiscordChannelImportFinishedSignal{
		CommunityID: communityID,
		ChannelID:   channelID,
	})
}

func SendDiscordCommunityImportCancelled(communityID string) {
	send(EventDiscordCommunityImportCancelled, DiscordCommunityImportCancelledSignal{
		CommunityID: communityID,
	})
}

func SendDiscordCommunityImportCleanedUp(communityID string) {
	send(EventDiscordCommunityImportCleanedUp, DiscordCommunityImportCleanedUpSignal{
		CommunityID: communityID,
	})
}

func SendDiscordChannelImportCancelled(channelID string) {
	send(EventDiscordChannelImportCancelled, DiscordChannelImportCancelledSignal{
		ChannelID: channelID,
	})
}
