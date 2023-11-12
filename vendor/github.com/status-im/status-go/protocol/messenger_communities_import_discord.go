package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/meirf/gopart"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/discord"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/protocol/transport"
	v1protocol "github.com/status-im/status-go/protocol/v1"
)

func (m *Messenger) ExtractDiscordDataFromImportFiles(filesToImport []string) (*discord.ExtractedData, map[string]*discord.ImportError) {

	extractedData := &discord.ExtractedData{
		Categories:             map[string]*discord.Category{},
		ExportedData:           make([]*discord.ExportedData, 0),
		OldestMessageTimestamp: 0,
		MessageCount:           0,
	}

	errors := map[string]*discord.ImportError{}

	for _, fileToImport := range filesToImport {
		filePath := strings.Replace(fileToImport, "file://", "", -1)

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			errors[fileToImport] = discord.Error(err.Error())
			continue
		}

		fileSize := fileInfo.Size()
		if fileSize > discord.MaxImportFileSizeBytes {
			errors[fileToImport] = discord.Error(discord.ErrImportFileTooBig.Error())
			continue
		}

		bytes, err := os.ReadFile(filePath)
		if err != nil {
			errors[fileToImport] = discord.Error(err.Error())
			continue
		}

		var discordExportedData discord.ExportedData

		err = json.Unmarshal(bytes, &discordExportedData)
		if err != nil {
			errors[fileToImport] = discord.Error(err.Error())
			continue
		}

		if len(discordExportedData.Messages) == 0 {
			errors[fileToImport] = discord.Error(discord.ErrNoMessageData.Error())
			continue
		}

		discordExportedData.Channel.FilePath = filePath
		categoryID := discordExportedData.Channel.CategoryID

		discordCategory := discord.Category{
			ID:   categoryID,
			Name: discordExportedData.Channel.CategoryName,
		}

		_, ok := extractedData.Categories[categoryID]
		if !ok {
			extractedData.Categories[categoryID] = &discordCategory
		}

		extractedData.MessageCount = extractedData.MessageCount + discordExportedData.MessageCount
		extractedData.ExportedData = append(extractedData.ExportedData, &discordExportedData)

		if len(discordExportedData.Messages) > 0 {
			msgTime, err := time.Parse(discordTimestampLayout, discordExportedData.Messages[0].Timestamp)
			if err != nil {
				m.logger.Error("failed to parse discord message timestamp", zap.Error(err))
				continue
			}

			if extractedData.OldestMessageTimestamp == 0 || int(msgTime.Unix()) <= extractedData.OldestMessageTimestamp {
				// Exported discord channel data already comes with `messages` being
				// sorted, starting with the oldest, so we can safely rely on the first
				// message
				extractedData.OldestMessageTimestamp = int(msgTime.Unix())
			}
		}
	}
	return extractedData, errors
}

func (m *Messenger) ExtractDiscordChannelsAndCategories(filesToImport []string) (*MessengerResponse, map[string]*discord.ImportError) {

	response := &MessengerResponse{}

	extractedData, errs := m.ExtractDiscordDataFromImportFiles(filesToImport)

	for _, category := range extractedData.Categories {
		response.AddDiscordCategory(category)
	}
	for _, export := range extractedData.ExportedData {
		response.AddDiscordChannel(&export.Channel)
	}
	if extractedData.OldestMessageTimestamp != 0 {
		response.DiscordOldestMessageTimestamp = extractedData.OldestMessageTimestamp
	}

	return response, errs
}

func (m *Messenger) RequestExtractDiscordChannelsAndCategories(filesToImport []string) {
	go func() {
		response, errors := m.ExtractDiscordChannelsAndCategories(filesToImport)
		m.config.messengerSignalsHandler.DiscordCategoriesAndChannelsExtracted(
			response.DiscordCategories,
			response.DiscordChannels,
			int64(response.DiscordOldestMessageTimestamp),
			errors)
	}()
}
func (m *Messenger) saveDiscordAuthorIfNotExists(discordAuthor *protobuf.DiscordMessageAuthor) *discord.ImportError {
	exists, err := m.persistence.HasDiscordMessageAuthor(discordAuthor.GetId())
	if err != nil {
		m.logger.Error("failed to check if message author exists in database", zap.Error(err))
		return discord.Error(err.Error())
	}

	if !exists {
		err := m.persistence.SaveDiscordMessageAuthor(discordAuthor)
		if err != nil {
			return discord.Error(err.Error())
		}
	}

	return nil
}

func (m *Messenger) convertDiscordMessageTimeStamp(discordMessage *protobuf.DiscordMessage, timestamp time.Time) *discord.ImportError {
	discordMessage.Timestamp = fmt.Sprintf("%d", timestamp.Unix())

	if discordMessage.TimestampEdited != "" {
		timestampEdited, err := time.Parse(discordTimestampLayout, discordMessage.TimestampEdited)
		if err != nil {
			m.logger.Error("failed to parse discord message timestamp", zap.Error(err))
			return discord.Warning(err.Error())
		}
		// Convert timestamp to unix timestamp
		discordMessage.TimestampEdited = fmt.Sprintf("%d", timestampEdited.Unix())
	}

	return nil
}

func (m *Messenger) createPinMessageFromDiscordMessage(message *common.Message, pinnedMessageID string, channelID string, community *communities.Community) (*common.PinMessage, *discord.ImportError) {
	pinMessage := protobuf.PinMessage{
		Clock:       message.WhisperTimestamp,
		MessageId:   pinnedMessageID,
		ChatId:      message.LocalChatID,
		MessageType: protobuf.MessageType_COMMUNITY_CHAT,
		Pinned:      true,
	}

	encodedPayload, err := proto.Marshal(&pinMessage)
	if err != nil {
		m.logger.Error("failed to parse marshal pin message", zap.Error(err))
		return nil, discord.Warning(err.Error())
	}

	wrappedPayload, err := v1protocol.WrapMessageV1(encodedPayload, protobuf.ApplicationMetadataMessage_PIN_MESSAGE, community.PrivateKey())
	if err != nil {
		m.logger.Error("failed to wrap pin message", zap.Error(err))
		return nil, discord.Warning(err.Error())
	}

	pinMessageToSave := &common.PinMessage{
		ID:               types.EncodeHex(v1protocol.MessageID(&community.PrivateKey().PublicKey, wrappedPayload)),
		PinMessage:       &pinMessage,
		LocalChatID:      channelID,
		From:             message.From,
		SigPubKey:        message.SigPubKey,
		WhisperTimestamp: message.WhisperTimestamp,
	}

	return pinMessageToSave, nil
}

func (m *Messenger) processDiscordMessages(discordChannel *discord.ExportedData,
	channel *Chat,
	importProgress *discord.ImportProgress,
	progressUpdates chan *discord.ImportProgress,
	fromDate int64,
	community *communities.Community) (
	map[string]*common.Message,
	[]*common.PinMessage,
	map[string]*protobuf.DiscordMessageAuthor,
	[]*protobuf.DiscordMessageAttachment) {

	messagesToSave := make(map[string]*common.Message, 0)
	pinMessagesToSave := make([]*common.PinMessage, 0)
	authorProfilesToSave := make(map[string]*protobuf.DiscordMessageAuthor, 0)
	messageAttachmentsToDownload := make([]*protobuf.DiscordMessageAttachment, 0)

	for _, discordMessage := range discordChannel.Messages {

		timestamp, err := time.Parse(discordTimestampLayout, discordMessage.Timestamp)
		if err != nil {
			m.logger.Error("failed to parse discord message timestamp", zap.Error(err))
			importProgress.AddTaskError(discord.ImportMessagesTask, discord.Warning(err.Error()))
			progressUpdates <- importProgress
			continue
		}

		if timestamp.Unix() < fromDate {
			progressUpdates <- importProgress
			continue
		}

		importErr := m.saveDiscordAuthorIfNotExists(discordMessage.Author)
		if importErr != nil {
			importProgress.AddTaskError(discord.ImportMessagesTask, importErr)
			progressUpdates <- importProgress
			continue
		}

		hasPayload, err := m.persistence.HasDiscordMessageAuthorImagePayload(discordMessage.Author.GetId())
		if err != nil {
			m.logger.Error("failed to check if message avatar payload exists in database", zap.Error(err))
			importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
			progressUpdates <- importProgress
			continue
		}

		if !hasPayload {
			authorProfilesToSave[discordMessage.Author.Id] = discordMessage.Author
		}

		// Convert timestamp to unix timestamp
		importErr = m.convertDiscordMessageTimeStamp(discordMessage, timestamp)
		if importErr != nil {
			importProgress.AddTaskError(discord.ImportMessagesTask, importErr)
			progressUpdates <- importProgress
			continue
		}

		for i := range discordMessage.Attachments {
			discordMessage.Attachments[i].MessageId = discordMessage.Id
		}
		messageAttachmentsToDownload = append(messageAttachmentsToDownload, discordMessage.Attachments...)

		clockAndTimestamp := uint64(timestamp.Unix()) * 1000
		communityPubKey := community.PrivateKey().PublicKey

		chatMessage := protobuf.ChatMessage{
			Timestamp:   clockAndTimestamp,
			MessageType: protobuf.MessageType_COMMUNITY_CHAT,
			ContentType: protobuf.ChatMessage_DISCORD_MESSAGE,
			Clock:       clockAndTimestamp,
			ChatId:      channel.ID,
			Payload: &protobuf.ChatMessage_DiscordMessage{
				DiscordMessage: discordMessage,
			},
		}

		// Handle message replies
		if discordMessage.Type == string(discord.MessageTypeReply) && discordMessage.Reference != nil {
			repliedMessageID := community.IDString() + discordMessage.Reference.MessageId
			if _, exists := messagesToSave[repliedMessageID]; exists {
				chatMessage.ResponseTo = repliedMessageID
			}
		}

		messageToSave := &common.Message{
			ID:               community.IDString() + discordMessage.Id,
			WhisperTimestamp: clockAndTimestamp,
			From:             types.EncodeHex(crypto.FromECDSAPub(&communityPubKey)),
			Seen:             true,
			LocalChatID:      channel.ID,
			SigPubKey:        &communityPubKey,
			CommunityID:      community.IDString(),
			ChatMessage:      &chatMessage,
		}

		err = messageToSave.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
		if err != nil {
			m.logger.Error("failed to prepare message content", zap.Error(err))
			importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
			progressUpdates <- importProgress
			continue
		}

		// Handle pin messages
		if discordMessage.Type == string(discord.MessageTypeChannelPinned) && discordMessage.Reference != nil {

			pinnedMessageID := community.IDString() + discordMessage.Reference.MessageId
			_, exists := messagesToSave[pinnedMessageID]
			if exists {
				pinMessageToSave, importErr := m.createPinMessageFromDiscordMessage(messageToSave, pinnedMessageID, channel.ID, community)
				if importErr != nil {
					importProgress.AddTaskError(discord.ImportMessagesTask, importErr)
					progressUpdates <- importProgress
					continue
				}

				pinMessagesToSave = append(pinMessagesToSave, pinMessageToSave)

				// Generate SystemMessagePinnedMessage
				systemMessage, importErr := m.generateSystemPinnedMessage(pinMessageToSave, channel, clockAndTimestamp, pinnedMessageID)
				if importErr != nil {
					importProgress.AddTaskError(discord.ImportMessagesTask, importErr)
					progressUpdates <- importProgress
					continue
				}

				messagesToSave[systemMessage.ID] = systemMessage
			}
		} else {
			messagesToSave[messageToSave.ID] = messageToSave
		}
	}

	return messagesToSave, pinMessagesToSave, authorProfilesToSave, messageAttachmentsToDownload
}

func calculateProgress(i int, t int, currentProgress float32) float32 {
	current := float32(1) / float32(t) * currentProgress
	if i > 1 {
		return float32(i-1)/float32(t) + current
	}
	return current
}

func (m *Messenger) MarkDiscordCommunityImportAsCancelled(communityID string) {
	m.importingCommunities[communityID] = true
}

func (m *Messenger) MarkDiscordChannelImportAsCancelled(channelID string) {
	m.importingChannels[channelID] = true
}

func (m *Messenger) DiscordImportMarkedAsCancelled(communityID string) bool {
	cancelled, exists := m.importingCommunities[communityID]
	return exists && cancelled
}

func (m *Messenger) DiscordImportChannelMarkedAsCancelled(channelID string) bool {
	cancelled, exists := m.importingChannels[channelID]
	return exists && cancelled
}

func (m *Messenger) cleanUpImports() {
	for id := range m.importingCommunities {
		m.cleanUpImport(id)
	}
}

func (m *Messenger) cleanUpImport(communityID string) {
	community, err := m.communitiesManager.GetByIDString(communityID)
	if err != nil {
		m.logger.Error("clean up failed, couldn't delete community", zap.Error(err))
		return
	}
	deleteErr := m.communitiesManager.DeleteCommunity(community.ID())
	if deleteErr != nil {
		m.logger.Error("clean up failed, couldn't delete community", zap.Error(deleteErr))
	}
	deleteErr = m.persistence.DeleteMessagesByCommunityID(community.IDString())
	if deleteErr != nil {
		m.logger.Error("clean up failed, couldn't delete community messages", zap.Error(deleteErr))
	}
	m.config.messengerSignalsHandler.DiscordCommunityImportCleanedUp(communityID)
}

func (m *Messenger) cleanUpImportChannel(communityID string, channelID string) {
	_, err := m.DeleteCommunityChat(types.HexBytes(communityID), channelID)
	if err != nil {
		m.logger.Error("clean up failed, couldn't delete community chat", zap.Error(err))
		return
	}

	err = m.persistence.DeleteMessagesByChatID(channelID)
	if err != nil {
		m.logger.Error("clean up failed, couldn't delete community chat messages", zap.Error(err))
		return
	}
}

func (m *Messenger) publishImportProgress(progress *discord.ImportProgress) {
	m.config.messengerSignalsHandler.DiscordCommunityImportProgress(progress)
}

func (m *Messenger) publishChannelImportProgress(progress *discord.ImportProgress) {
	m.config.messengerSignalsHandler.DiscordChannelImportProgress(progress)
}

func (m *Messenger) startPublishImportProgressInterval(c chan *discord.ImportProgress, cancel chan string, done chan struct{}) {

	var currentProgress *discord.ImportProgress

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if currentProgress != nil {
					m.publishImportProgress(currentProgress)
					if currentProgress.Stopped {
						return
					}
				}
			case progressUpdate := <-c:
				currentProgress = progressUpdate
			case <-done:
				if currentProgress != nil {
					m.publishImportProgress(currentProgress)
				}
				return
			case communityID := <-cancel:
				if currentProgress != nil {
					m.publishImportProgress(currentProgress)
				}
				m.cleanUpImport(communityID)
				m.config.messengerSignalsHandler.DiscordCommunityImportCancelled(communityID)
				return
			case <-m.quit:
				m.cleanUpImports()
				return
			}
		}
	}()
}

func (m *Messenger) startPublishImportChannelProgressInterval(c chan *discord.ImportProgress, cancel chan []string, done chan struct{}) {

	var currentProgress *discord.ImportProgress

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if currentProgress != nil {
					m.publishChannelImportProgress(currentProgress)
					if currentProgress.Stopped {
						return
					}
				}
			case progressUpdate := <-c:
				currentProgress = progressUpdate
			case <-done:
				if currentProgress != nil {
					m.publishChannelImportProgress(currentProgress)
				}
				return
			case ids := <-cancel:
				if currentProgress != nil {
					m.publishImportProgress(currentProgress)
				}
				if len(ids) > 0 {
					communityID := ids[0]
					channelID := ids[1]
					discordChannelID := ids[2]
					m.cleanUpImportChannel(communityID, channelID)
					m.config.messengerSignalsHandler.DiscordChannelImportCancelled(discordChannelID)
				}
				return
			case <-m.quit:
				m.cleanUpImports()
				return
			}
		}
	}()
}
func createCommunityChannelForImport(request *requests.ImportDiscordChannel) *protobuf.CommunityChat {
	return &protobuf.CommunityChat{
		Permissions: &protobuf.CommunityPermissions{
			Access: protobuf.CommunityPermissions_AUTO_ACCEPT,
		},
		Identity: &protobuf.ChatIdentity{
			DisplayName: request.Name,
			Emoji:       request.Emoji,
			Description: request.Description,
			Color:       request.Color,
		},
		CategoryId: "",
	}
}

func (m *Messenger) RequestImportDiscordChannel(request *requests.ImportDiscordChannel) {
	go func() {
		totalImportChunkCount := len(request.FilesToImport)

		progressUpdates := make(chan *discord.ImportProgress)

		done := make(chan struct{})
		cancel := make(chan []string)

		var newChat *Chat

		m.startPublishImportChannelProgressInterval(progressUpdates, cancel, done)

		importProgress := &discord.ImportProgress{}
		importProgress.Init(totalImportChunkCount, []discord.ImportTask{
			discord.ChannelsCreationTask,
			discord.ImportMessagesTask,
			discord.DownloadAssetsTask,
			discord.InitCommunityTask,
		})

		importProgress.ChannelID = request.DiscordChannelID
		importProgress.ChannelName = request.Name
		// initial progress immediately

		if err := request.Validate(); err != nil {
			errmsg := fmt.Sprintf("Request validation failed: '%s'", err.Error())
			importProgress.AddTaskError(discord.ChannelsCreationTask, discord.Error(errmsg))
			importProgress.StopTask(discord.ChannelsCreationTask)
			progressUpdates <- importProgress
			cancel <- []string{request.CommunityID.String(), "", request.DiscordChannelID}
			return
		}

		// Here's 3 steps: Find the corrent channel in files, get the community and create the channel
		progressValue := float32(0.3)

		m.publishChannelImportProgress(importProgress)

		community, err := m.GetCommunityByID(request.CommunityID)

		if err != nil {
			errmsg := fmt.Sprintf("Couldn't get the community '%s': '%s'", request.CommunityID, err.Error())
			importProgress.AddTaskError(discord.ChannelsCreationTask, discord.Error(errmsg))
			importProgress.StopTask(discord.ChannelsCreationTask)
			progressUpdates <- importProgress
			cancel <- []string{request.CommunityID.String(), "", request.DiscordChannelID}
			return
		}

		importProgress.UpdateTaskProgress(discord.ChannelsCreationTask, progressValue)
		progressUpdates <- importProgress

		for i, importFile := range request.FilesToImport {
			m.importingChannels[request.DiscordChannelID] = false

			exportData, errs := m.ExtractDiscordDataFromImportFiles([]string{importFile})
			if len(errs) > 0 {
				for _, err := range errs {
					importProgress.AddTaskError(discord.ChannelsCreationTask, err)
				}
				importProgress.StopTask(discord.ChannelsCreationTask)
				progressUpdates <- importProgress
				cancel <- []string{request.CommunityID.String(), "", request.DiscordChannelID}
				return
			}

			var channel *discord.ExportedData

			for _, ch := range exportData.ExportedData {
				if ch.Channel.ID == request.DiscordChannelID {
					channel = ch
				}
			}

			if channel == nil {
				if i < len(request.FilesToImport)-1 {
					// skip this file
					continue
				} else if i == len(request.FilesToImport)-1 {
					errmsg := fmt.Sprintf("Couldn't find the target channel id in files: '%s'", request.DiscordChannelID)
					importProgress.AddTaskError(discord.ChannelsCreationTask, discord.Error(errmsg))
					importProgress.StopTask(discord.ChannelsCreationTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), "", request.DiscordChannelID}
					return
				}
			}
			progressValue := float32(0.6)

			importProgress.UpdateTaskProgress(discord.ChannelsCreationTask, progressValue)
			progressUpdates <- importProgress

			if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
				importProgress.StopTask(discord.ChannelsCreationTask)
				progressUpdates <- importProgress
				cancel <- []string{request.CommunityID.String(), "", request.DiscordChannelID}
				return
			}

			if len(channel.Channel.ID) == 0 {
				// skip this file and try to find in the next file
				continue
			}
			exists := false

			for _, chatID := range community.ChatIDs() {
				if strings.HasSuffix(chatID, request.DiscordChannelID) {
					exists = true
					break
				}
			}

			if !exists {
				communityChat := createCommunityChannelForImport(request)

				changes, err := m.communitiesManager.CreateChat(request.CommunityID, communityChat, false, channel.Channel.ID)
				if err != nil {
					errmsg := err.Error()
					if errors.Is(err, communities.ErrInvalidCommunityDescriptionDuplicatedName) {
						errmsg = fmt.Sprintf("Couldn't create channel '%s': %s", communityChat.Identity.DisplayName, err.Error())
						fmt.Println(errmsg)
					}

					importProgress.AddTaskError(discord.ChannelsCreationTask, discord.Error(errmsg))
					importProgress.StopTask(discord.ChannelsCreationTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), "", request.DiscordChannelID}
					return
				}

				community = changes.Community
				for chatID, chat := range changes.ChatsAdded {
					newChat = CreateCommunityChat(request.CommunityID.String(), chatID, chat, m.getTimesource())
				}

				progressValue = float32(1.0)

				importProgress.UpdateTaskProgress(discord.ChannelsCreationTask, progressValue)
				progressUpdates <- importProgress
			} else {
				// When channel with current discord id already exist we should skip import
				importProgress.AddTaskError(discord.ChannelsCreationTask, discord.Error("Channel already imported to this community"))
				importProgress.StopTask(discord.ChannelsCreationTask)
				progressUpdates <- importProgress
				cancel <- []string{request.CommunityID.String(), "", request.DiscordChannelID}
				return
			}

			if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
				importProgress.StopTask(discord.ImportMessagesTask)
				progressUpdates <- importProgress
				cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
				return
			}

			messagesToSave, pinMessagesToSave, authorProfilesToSave, messageAttachmentsToDownload :=
				m.processDiscordMessages(channel, newChat, importProgress, progressUpdates, request.From, community)

			// If either there were no messages in the channel or something happened and all the messages errored, we
			// we still up the percent to 100%
			if len(messagesToSave) == 0 {
				importProgress.UpdateTaskProgress(discord.ImportMessagesTask, 1.0)
				progressUpdates <- importProgress
			}

			var discordMessages []*protobuf.DiscordMessage
			for _, msg := range messagesToSave {
				if msg.ChatMessage.ContentType == protobuf.ChatMessage_DISCORD_MESSAGE {
					discordMessages = append(discordMessages, msg.GetDiscordMessage())
				}
			}

			// We save these messages in chunks, so we don't block the database
			// for a longer period of time
			discordMessageChunks := chunkSlice(discordMessages, maxChunkSizeMessages)
			chunksCount := len(discordMessageChunks)

			for ii, msgs := range discordMessageChunks {
				m.communitiesManager.LogStdout(fmt.Sprintf("saving %d/%d chunk with %d discord messages", ii+1, chunksCount, len(msgs)))
				err := m.persistence.SaveDiscordMessages(msgs)
				if err != nil {
					m.cleanUpImportChannel(request.CommunityID.String(), newChat.ID)
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
					return
				}

				if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
					return
				}

				// We're multiplying `chunksCount` by `0.25` so we leave 25% for additional save operations
				// 0.5 are the previous 50% of progress
				currentCount := ii + 1
				progressValue := calculateProgress(i+1, totalImportChunkCount, 0.5+(float32(currentCount)/float32(chunksCount))*0.25)
				importProgress.UpdateTaskProgress(discord.ImportMessagesTask, progressValue)
				progressUpdates <- importProgress

				// We slow down the saving of message chunks to keep the database responsive
				if currentCount < chunksCount {
					time.Sleep(2 * time.Second)
				}
			}

			// Get slice of all values in `messagesToSave` map
			var messages = make([]*common.Message, 0, len(messagesToSave))
			for _, msg := range messagesToSave {
				messages = append(messages, msg)
			}

			// Same as above, we save these messages in chunks so we don't block
			// the database for a longer period of time
			messageChunks := chunkSlice(messages, maxChunkSizeMessages)
			chunksCount = len(messageChunks)

			for ii, msgs := range messageChunks {
				m.communitiesManager.LogStdout(fmt.Sprintf("saving %d/%d chunk with %d app messages", ii+1, chunksCount, len(msgs)))
				err := m.persistence.SaveMessages(msgs)
				if err != nil {
					m.cleanUpImportChannel(request.CommunityID.String(), request.DiscordChannelID)
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}

					return
				}

				if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
					return
				}

				// 0.75 are the previous 75% of progress, hence we multiply our chunk progress
				// by 0.25
				currentCount := ii + 1
				progressValue := calculateProgress(i+1, totalImportChunkCount, 0.75+(float32(currentCount)/float32(chunksCount))*0.25)
				// progressValue := 0.75 + ((float32(currentCount) / float32(chunksCount)) * 0.25)
				importProgress.UpdateTaskProgress(discord.ImportMessagesTask, progressValue)
				progressUpdates <- importProgress

				// We slow down the saving of message chunks to keep the database responsive
				if currentCount < chunksCount {
					time.Sleep(2 * time.Second)
				}
			}

			pinMessageChunks := chunkSlice(pinMessagesToSave, maxChunkSizeMessages)
			for _, pinMsgs := range pinMessageChunks {
				err := m.persistence.SavePinMessages(pinMsgs)
				if err != nil {
					m.cleanUpImportChannel(request.CommunityID.String(), request.DiscordChannelID)
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}

					return
				}

				if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
					return
				}
			}

			totalAssetsCount := len(messageAttachmentsToDownload) + len(authorProfilesToSave)
			var assetCounter discord.AssetCounter

			var wg sync.WaitGroup

			for id, author := range authorProfilesToSave {
				wg.Add(1)
				go func(id string, author *protobuf.DiscordMessageAuthor) {
					defer wg.Done()

					m.communitiesManager.LogStdout(fmt.Sprintf("downloading asset %d/%d", assetCounter.Value()+1, totalAssetsCount))
					imagePayload, err := discord.DownloadAvatarAsset(author.AvatarUrl)
					if err != nil {
						errmsg := fmt.Sprintf("Couldn't download profile avatar '%s': %s", author.AvatarUrl, err.Error())
						importProgress.AddTaskError(
							discord.DownloadAssetsTask,
							discord.Warning(errmsg),
						)
						progressUpdates <- importProgress

						return
					}

					err = m.persistence.UpdateDiscordMessageAuthorImage(author.Id, imagePayload)
					if err != nil {
						importProgress.AddTaskError(discord.DownloadAssetsTask, discord.Warning(err.Error()))
						progressUpdates <- importProgress

						return
					}

					author.AvatarImagePayload = imagePayload
					authorProfilesToSave[id] = author

					if m.DiscordImportMarkedAsCancelled(request.DiscordChannelID) {
						importProgress.StopTask(discord.DownloadAssetsTask)
						progressUpdates <- importProgress
						cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
						return
					}

					assetCounter.Increase()
					progressValue := calculateProgress(i+1, totalImportChunkCount, (float32(assetCounter.Value())/float32(totalAssetsCount))*0.5)
					importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
					progressUpdates <- importProgress

				}(id, author)
			}
			wg.Wait()

			if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
				importProgress.StopTask(discord.DownloadAssetsTask)
				progressUpdates <- importProgress
				cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
				return
			}

			for idxRange := range gopart.Partition(len(messageAttachmentsToDownload), 100) {
				attachments := messageAttachmentsToDownload[idxRange.Low:idxRange.High]
				wg.Add(1)
				go func(attachments []*protobuf.DiscordMessageAttachment) {
					defer wg.Done()
					for ii, attachment := range attachments {

						m.communitiesManager.LogStdout(fmt.Sprintf("downloading asset %d/%d", assetCounter.Value()+1, totalAssetsCount))

						assetPayload, contentType, err := discord.DownloadAsset(attachment.Url)
						if err != nil {
							errmsg := fmt.Sprintf("Couldn't download message attachment '%s': %s", attachment.Url, err.Error())
							importProgress.AddTaskError(
								discord.DownloadAssetsTask,
								discord.Warning(errmsg),
							)
							progressUpdates <- importProgress
							continue
						}

						attachment.Payload = assetPayload
						attachment.ContentType = contentType
						messageAttachmentsToDownload[ii] = attachment

						if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
							importProgress.StopTask(discord.DownloadAssetsTask)
							progressUpdates <- importProgress
							cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
							return
						}

						assetCounter.Increase()
						progressValue := calculateProgress(i+1, totalImportChunkCount, (float32(assetCounter.Value())/float32(totalAssetsCount))*0.5)
						importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
						progressUpdates <- importProgress
					}
				}(attachments)
			}
			wg.Wait()

			if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
				importProgress.StopTask(discord.DownloadAssetsTask)
				progressUpdates <- importProgress
				cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
				return
			}

			attachmentChunks := chunkAttachmentsByByteSize(messageAttachmentsToDownload, maxChunkSizeBytes)
			chunksCount = len(attachmentChunks)

			for ii, attachments := range attachmentChunks {
				m.communitiesManager.LogStdout(fmt.Sprintf("saving %d/%d chunk with %d discord message attachments", ii+1, chunksCount, len(attachments)))
				err := m.persistence.SaveDiscordMessageAttachments(attachments)
				if err != nil {
					importProgress.AddTaskError(discord.DownloadAssetsTask, discord.Warning(err.Error()))
					importProgress.Stop()
					progressUpdates <- importProgress

					continue
				}

				if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
					importProgress.StopTask(discord.DownloadAssetsTask)
					progressUpdates <- importProgress
					cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
					return
				}

				// 0.5 are the previous 50% of progress, hence we multiply our chunk progress
				// by 0.5
				currentCount := ii + 1
				progressValue := calculateProgress(i+1, totalImportChunkCount, 0.5+(float32(currentCount)/float32(chunksCount))*0.5)
				importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
				progressUpdates <- importProgress

				// We slow down the saving of attachment chunks to keep the database responsive
				if currentCount < chunksCount {
					time.Sleep(2 * time.Second)
				}
			}

			if len(attachmentChunks) == 0 {
				progressValue := calculateProgress(i+1, totalImportChunkCount, 1.0)
				importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
			}

			_, err := m.transport.JoinPublic(newChat.ID)
			if err != nil {
				m.logger.Error("failed to load filter for chat", zap.Error(err))
				continue
			}

			wakuChatMessages, err := m.chatMessagesToWakuMessages(messages, community)
			if err != nil {
				m.logger.Error("failed to convert chat messages into waku messages", zap.Error(err))
				continue
			}

			wakuPinMessages, err := m.pinMessagesToWakuMessages(pinMessagesToSave, community)
			if err != nil {
				m.logger.Error("failed to convert pin messages into waku messages", zap.Error(err))
				continue
			}

			wakuMessages := append(wakuChatMessages, wakuPinMessages...)

			topics, err := m.communitiesManager.GetCommunityChatsTopics(request.CommunityID)
			if err != nil {
				m.logger.Error("failed to get community chat topics", zap.Error(err))
				continue
			}

			startDate := time.Unix(int64(exportData.OldestMessageTimestamp), 0)
			endDate := time.Now()

			_, err = m.communitiesManager.CreateHistoryArchiveTorrentFromMessages(
				request.CommunityID,
				wakuMessages,
				topics,
				startDate,
				endDate,
				messageArchiveInterval,
				community.Encrypted(),
			)
			if err != nil {
				m.logger.Error("failed to create history archive torrent", zap.Error(err))
				continue
			}
			communitySettings, err := m.communitiesManager.GetCommunitySettingsByID(request.CommunityID)
			if err != nil {
				m.logger.Error("Failed to get community settings", zap.Error(err))
				continue
			}
			if m.torrentClientReady() && communitySettings.HistoryArchiveSupportEnabled {

				err = m.communitiesManager.SeedHistoryArchiveTorrent(request.CommunityID)
				if err != nil {
					m.logger.Error("failed to seed history archive", zap.Error(err))
				}
				go m.communitiesManager.StartHistoryArchiveTasksInterval(community, messageArchiveInterval)
			}
		}

		importProgress.UpdateTaskProgress(discord.InitCommunityTask, float32(0.0))

		if m.DiscordImportChannelMarkedAsCancelled(request.DiscordChannelID) {
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			cancel <- []string{request.CommunityID.String(), newChat.ID, request.DiscordChannelID}
			return
		}

		// Chats need to be saved after the community has been published,
		// hence we make this part of the `InitCommunityTask`
		err = m.saveChat(newChat)

		if err != nil {
			m.cleanUpImportChannel(request.CommunityID.String(), request.DiscordChannelID)
			importProgress.AddTaskError(discord.InitCommunityTask, discord.Error(err.Error()))
			importProgress.Stop()
			progressUpdates <- importProgress
			cancel <- []string{request.CommunityID.String(), request.DiscordChannelID}
			return
		}

		// Make sure all progress tasks are at 100%, in case one of the steps had errors
		// The front-end doesn't understand that the import is done until all tasks are at 100%
		importProgress.UpdateTaskProgress(discord.CommunityCreationTask, float32(1.0))
		importProgress.UpdateTaskProgress(discord.ChannelsCreationTask, float32(1.0))
		importProgress.UpdateTaskProgress(discord.ImportMessagesTask, float32(1.0))
		importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, float32(1.0))
		importProgress.UpdateTaskProgress(discord.InitCommunityTask, float32(1.0))

		m.config.messengerSignalsHandler.DiscordChannelImportFinished(request.CommunityID.String(), newChat.ID)
		close(done)
	}()
}

func (m *Messenger) RequestImportDiscordCommunity(request *requests.ImportDiscordCommunity) {
	go func() {

		totalImportChunkCount := len(request.FilesToImport)

		progressUpdates := make(chan *discord.ImportProgress)
		done := make(chan struct{})
		cancel := make(chan string)
		m.startPublishImportProgressInterval(progressUpdates, cancel, done)

		importProgress := &discord.ImportProgress{}
		importProgress.Init(totalImportChunkCount, []discord.ImportTask{
			discord.CommunityCreationTask,
			discord.ChannelsCreationTask,
			discord.ImportMessagesTask,
			discord.DownloadAssetsTask,
			discord.InitCommunityTask,
		})
		importProgress.CommunityName = request.Name

		// initial progress immediately
		m.publishImportProgress(importProgress)

		createCommunityRequest := request.ToCreateCommunityRequest()

		// We're calling `CreateCommunity` on `communitiesManager` directly, instead of
		// using the `Messenger` API, so we get more control over when we set up filters,
		// the community is published and data is being synced (we don't want the community
		// to show up in clients while the import is in progress)
		discordCommunity, err := m.communitiesManager.CreateCommunity(createCommunityRequest, false)
		if err != nil {
			importProgress.AddTaskError(discord.CommunityCreationTask, discord.Error(err.Error()))
			importProgress.StopTask(discord.CommunityCreationTask)
			progressUpdates <- importProgress
			return
		}

		communitySettings := communities.CommunitySettings{
			CommunityID:                  discordCommunity.IDString(),
			HistoryArchiveSupportEnabled: true,
		}
		err = m.communitiesManager.SaveCommunitySettings(communitySettings)
		if err != nil {
			m.cleanUpImport(discordCommunity.IDString())
			importProgress.AddTaskError(discord.CommunityCreationTask, discord.Error(err.Error()))
			importProgress.StopTask(discord.CommunityCreationTask)
			progressUpdates <- importProgress
			return
		}

		communityID := discordCommunity.IDString()

		// marking import as not cancelled
		m.importingCommunities[communityID] = false
		importProgress.CommunityID = communityID
		importProgress.CommunityImages = make(map[string]images.IdentityImage)

		imgs := discordCommunity.Images()
		for t, i := range imgs {
			importProgress.CommunityImages[t] = images.IdentityImage{Name: t, Payload: i.Payload}
		}

		importProgress.UpdateTaskProgress(discord.CommunityCreationTask, 1)
		progressUpdates <- importProgress

		if m.DiscordImportMarkedAsCancelled(communityID) {
			importProgress.StopTask(discord.CommunityCreationTask)
			progressUpdates <- importProgress
			cancel <- communityID
			return
		}

		var chatsToSave []*Chat
		createdChats := make(map[string]*Chat, 0)
		processedChannelIds := make(map[string]string, 0)
		processedCategoriesIds := make(map[string]string, 0)

		// The map with counts of duplicated channel names
		uniqueChatNames := make(map[string]int, 0)

		for i, importFile := range request.FilesToImport {

			exportData, errs := m.ExtractDiscordDataFromImportFiles([]string{importFile})
			if len(errs) > 0 {
				for _, err := range errs {
					importProgress.AddTaskError(discord.CommunityCreationTask, err)
				}
				progressUpdates <- importProgress
				return
			}
			totalChannelsCount := len(exportData.ExportedData)
			totalMessageCount := exportData.MessageCount

			if totalChannelsCount == 0 || totalMessageCount == 0 {
				importError := discord.Error(fmt.Errorf("No channel to import messages from in file: %s", importFile).Error())
				if totalMessageCount == 0 {
					importError.Message = fmt.Errorf("No messages to import in file: %s", importFile).Error()
				}
				importProgress.AddTaskError(discord.ChannelsCreationTask, importError)
				progressUpdates <- importProgress
				continue
			}

			importProgress.CurrentChunk = i + 1

			// We actually only ever receive a single category
			// from `exportData` but since it's a map, we still have to
			// iterate over it to access its values
			for _, category := range exportData.Categories {

				categories := discordCommunity.Categories()
				exists := false
				for catID := range categories {
					if strings.HasSuffix(catID, category.ID) {
						exists = true
						break
					}
				}

				if !exists {
					createCommunityCategoryRequest := &requests.CreateCommunityCategory{
						CommunityID:  discordCommunity.ID(),
						CategoryName: category.Name,
						ThirdPartyID: category.ID,
						ChatIDs:      make([]string, 0),
					}
					// We call `CreateCategory` on `communitiesManager` directly so we can control
					// whether or not the community update should be published (it should not until the
					// import has finished)
					communityWithCategories, changes, err := m.communitiesManager.CreateCategory(createCommunityCategoryRequest, false)
					if err != nil {
						m.cleanUpImport(communityID)
						importProgress.AddTaskError(discord.CommunityCreationTask, discord.Error(err.Error()))
						importProgress.StopTask(discord.CommunityCreationTask)
						progressUpdates <- importProgress
						return
					}
					discordCommunity = communityWithCategories
					// This looks like we keep overriding the same field but there's
					// only one `CategoriesAdded` change at this point.
					for _, addedCategory := range changes.CategoriesAdded {
						processedCategoriesIds[category.ID] = addedCategory.CategoryId
					}
				}
			}

			progressValue := calculateProgress(i+1, totalImportChunkCount, (float32(1) / 2))
			importProgress.UpdateTaskProgress(discord.ChannelsCreationTask, progressValue)

			progressUpdates <- importProgress

			if m.DiscordImportMarkedAsCancelled(communityID) {
				importProgress.StopTask(discord.CommunityCreationTask)
				progressUpdates <- importProgress
				cancel <- communityID
				return
			}

			messagesToSave := make(map[string]*common.Message, 0)
			pinMessagesToSave := make([]*common.PinMessage, 0)
			authorProfilesToSave := make(map[string]*protobuf.DiscordMessageAuthor, 0)
			messageAttachmentsToDownload := make([]*protobuf.DiscordMessageAttachment, 0)

			// Save to access the first item here as we process
			// exported data by files which only holds a single channel
			channel := exportData.ExportedData[0]
			chatIDs := discordCommunity.ChatIDs()

			exists := false
			for _, chatID := range chatIDs {
				if strings.HasSuffix(chatID, channel.Channel.ID) {
					exists = true
					break
				}
			}

			if !exists {
				channelUniqueName := channel.Channel.Name
				if count, ok := uniqueChatNames[channelUniqueName]; ok {
					uniqueChatNames[channelUniqueName] = count + 1
					channelUniqueName = fmt.Sprintf("%s_%d", channelUniqueName, uniqueChatNames[channelUniqueName])
				} else {
					uniqueChatNames[channelUniqueName] = 1
				}

				communityChat := &protobuf.CommunityChat{
					Permissions: &protobuf.CommunityPermissions{
						Access: protobuf.CommunityPermissions_AUTO_ACCEPT,
					},
					Identity: &protobuf.ChatIdentity{
						DisplayName: channelUniqueName,
						Emoji:       "",
						Description: channel.Channel.Description,
						Color:       discordCommunity.Color(),
					},
					CategoryId: processedCategoriesIds[channel.Channel.CategoryID],
				}

				// We call `CreateChat` on `communitiesManager` directly to get more control
				// over whether we want to publish the updated community description.
				changes, err := m.communitiesManager.CreateChat(discordCommunity.ID(), communityChat, false, channel.Channel.ID)
				if err != nil {
					m.cleanUpImport(communityID)
					errmsg := err.Error()
					if errors.Is(err, communities.ErrInvalidCommunityDescriptionDuplicatedName) {
						errmsg = fmt.Sprintf("Couldn't create channel '%s': %s", communityChat.Identity.DisplayName, err.Error())
					}
					importProgress.AddTaskError(discord.ChannelsCreationTask, discord.Error(errmsg))
					importProgress.StopTask(discord.ChannelsCreationTask)
					progressUpdates <- importProgress
					return
				}
				discordCommunity = changes.Community

				// This looks like we keep overriding the chat id value
				// as we iterate over `ChatsAdded`, however at this point we
				// know there was only a single such change (and it's a map)
				for chatID, chat := range changes.ChatsAdded {
					c := CreateCommunityChat(communityID, chatID, chat, m.getTimesource())
					createdChats[c.ID] = c
					chatsToSave = append(chatsToSave, c)
					processedChannelIds[channel.Channel.ID] = c.ID
				}
			}

			progressValue = calculateProgress(i+1, totalImportChunkCount, 1)
			importProgress.UpdateTaskProgress(discord.ChannelsCreationTask, progressValue)
			progressUpdates <- importProgress

			for ii, discordMessage := range channel.Messages {

				timestamp, err := time.Parse(discordTimestampLayout, discordMessage.Timestamp)
				if err != nil {
					m.logger.Error("failed to parse discord message timestamp", zap.Error(err))
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Warning(err.Error()))
					progressUpdates <- importProgress
					continue
				}

				if timestamp.Unix() < request.From {
					progressUpdates <- importProgress
					continue
				}

				exists, err := m.persistence.HasDiscordMessageAuthor(discordMessage.Author.GetId())
				if err != nil {
					m.logger.Error("failed to check if message author exists in database", zap.Error(err))
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					progressUpdates <- importProgress
					continue
				}

				if !exists {
					err := m.persistence.SaveDiscordMessageAuthor(discordMessage.Author)
					if err != nil {
						importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
						progressUpdates <- importProgress
						continue
					}
				}

				hasPayload, err := m.persistence.HasDiscordMessageAuthorImagePayload(discordMessage.Author.GetId())
				if err != nil {
					m.logger.Error("failed to check if message avatar payload exists in database", zap.Error(err))
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					progressUpdates <- importProgress
					continue
				}

				if !hasPayload {
					authorProfilesToSave[discordMessage.Author.Id] = discordMessage.Author
				}

				// Convert timestamp to unix timestamp
				discordMessage.Timestamp = fmt.Sprintf("%d", timestamp.Unix())

				if discordMessage.TimestampEdited != "" {
					timestampEdited, err := time.Parse(discordTimestampLayout, discordMessage.TimestampEdited)
					if err != nil {
						m.logger.Error("failed to parse discord message timestamp", zap.Error(err))
						importProgress.AddTaskError(discord.ImportMessagesTask, discord.Warning(err.Error()))
						progressUpdates <- importProgress
						continue
					}
					// Convert timestamp to unix timestamp
					discordMessage.TimestampEdited = fmt.Sprintf("%d", timestampEdited.Unix())
				}

				for i := range discordMessage.Attachments {
					discordMessage.Attachments[i].MessageId = discordMessage.Id
				}
				messageAttachmentsToDownload = append(messageAttachmentsToDownload, discordMessage.Attachments...)

				clockAndTimestamp := uint64(timestamp.Unix()) * 1000
				communityPubKey := discordCommunity.PrivateKey().PublicKey

				chatMessage := protobuf.ChatMessage{
					Timestamp:   clockAndTimestamp,
					MessageType: protobuf.MessageType_COMMUNITY_CHAT,
					ContentType: protobuf.ChatMessage_DISCORD_MESSAGE,
					Clock:       clockAndTimestamp,
					ChatId:      processedChannelIds[channel.Channel.ID],
					Payload: &protobuf.ChatMessage_DiscordMessage{
						DiscordMessage: discordMessage,
					},
				}

				// Handle message replies
				if discordMessage.Type == string(discord.MessageTypeReply) && discordMessage.Reference != nil {
					repliedMessageID := communityID + discordMessage.Reference.MessageId
					if _, exists := messagesToSave[repliedMessageID]; exists {
						chatMessage.ResponseTo = repliedMessageID
					}
				}

				messageToSave := &common.Message{
					ID:               communityID + discordMessage.Id,
					WhisperTimestamp: clockAndTimestamp,
					From:             types.EncodeHex(crypto.FromECDSAPub(&communityPubKey)),
					Seen:             true,
					LocalChatID:      processedChannelIds[channel.Channel.ID],
					SigPubKey:        &communityPubKey,
					CommunityID:      communityID,
					ChatMessage:      &chatMessage,
				}

				err = messageToSave.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
				if err != nil {
					m.logger.Error("failed to prepare message content", zap.Error(err))
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					progressUpdates <- importProgress
					continue
				}

				// Handle pin messages
				if discordMessage.Type == string(discord.MessageTypeChannelPinned) && discordMessage.Reference != nil {

					pinnedMessageID := communityID + discordMessage.Reference.MessageId
					_, exists := messagesToSave[pinnedMessageID]
					if exists {
						pinMessage := protobuf.PinMessage{
							Clock:       messageToSave.WhisperTimestamp,
							MessageId:   pinnedMessageID,
							ChatId:      messageToSave.LocalChatID,
							MessageType: protobuf.MessageType_COMMUNITY_CHAT,
							Pinned:      true,
						}

						encodedPayload, err := proto.Marshal(&pinMessage)
						if err != nil {
							m.logger.Error("failed to parse marshal pin message", zap.Error(err))
							importProgress.AddTaskError(discord.ImportMessagesTask, discord.Warning(err.Error()))
							progressUpdates <- importProgress
							continue
						}

						wrappedPayload, err := v1protocol.WrapMessageV1(encodedPayload, protobuf.ApplicationMetadataMessage_PIN_MESSAGE, discordCommunity.PrivateKey())
						if err != nil {
							m.logger.Error("failed to wrap pin message", zap.Error(err))
							importProgress.AddTaskError(discord.ImportMessagesTask, discord.Warning(err.Error()))
							progressUpdates <- importProgress
							continue
						}

						pinMessageToSave := common.PinMessage{
							ID:               types.EncodeHex(v1protocol.MessageID(&communityPubKey, wrappedPayload)),
							PinMessage:       &pinMessage,
							LocalChatID:      processedChannelIds[channel.Channel.ID],
							From:             messageToSave.From,
							SigPubKey:        messageToSave.SigPubKey,
							WhisperTimestamp: messageToSave.WhisperTimestamp,
						}

						pinMessagesToSave = append(pinMessagesToSave, &pinMessageToSave)

						// Generate SystemMessagePinnedMessage

						chat, ok := createdChats[pinMessageToSave.LocalChatID]
						if !ok {
							err := errors.New("failed to get chat for pin message")
							m.logger.Warn(err.Error(),
								zap.String("PinMessageId", pinMessageToSave.ID),
								zap.String("ChatID", pinMessageToSave.LocalChatID))
							importProgress.AddTaskError(discord.ImportMessagesTask, discord.Warning(err.Error()))
							progressUpdates <- importProgress
							continue
						}

						id, err := generatePinMessageNotificationID(&m.identity.PublicKey, &pinMessageToSave, chat)
						if err != nil {
							m.logger.Warn("failed to generate pin message notification ID",
								zap.String("PinMessageId", pinMessageToSave.ID))
							importProgress.AddTaskError(discord.ImportMessagesTask, discord.Warning(err.Error()))
							progressUpdates <- importProgress
							continue
						}
						systemMessage := &common.Message{
							ChatMessage: &protobuf.ChatMessage{
								Clock:       pinMessageToSave.Clock,
								Timestamp:   clockAndTimestamp,
								ChatId:      chat.ID,
								MessageType: pinMessageToSave.MessageType,
								ResponseTo:  pinMessage.MessageId,
								ContentType: protobuf.ChatMessage_SYSTEM_MESSAGE_PINNED_MESSAGE,
							},
							WhisperTimestamp: clockAndTimestamp,
							ID:               id,
							LocalChatID:      chat.ID,
							From:             messageToSave.From,
							Seen:             true,
						}

						messagesToSave[systemMessage.ID] = systemMessage
					}
				} else {
					messagesToSave[messageToSave.ID] = messageToSave
				}

				progressValue := calculateProgress(i+1, totalImportChunkCount, float32(ii+1)/float32(len(channel.Messages))*0.5)
				importProgress.UpdateTaskProgress(discord.ImportMessagesTask, progressValue)
				progressUpdates <- importProgress
			}

			if m.DiscordImportMarkedAsCancelled(communityID) {
				importProgress.StopTask(discord.ImportMessagesTask)
				progressUpdates <- importProgress
				cancel <- communityID
				return
			}

			var discordMessages []*protobuf.DiscordMessage
			for _, msg := range messagesToSave {
				if msg.ChatMessage.ContentType == protobuf.ChatMessage_DISCORD_MESSAGE {
					discordMessages = append(discordMessages, msg.GetDiscordMessage())
				}
			}

			// We save these messages in chunks, so we don't block the database
			// for a longer period of time
			discordMessageChunks := chunkSlice(discordMessages, maxChunkSizeMessages)
			chunksCount := len(discordMessageChunks)

			for ii, msgs := range discordMessageChunks {
				m.communitiesManager.LogStdout(fmt.Sprintf("saving %d/%d chunk with %d discord messages", ii+1, chunksCount, len(msgs)))
				err = m.persistence.SaveDiscordMessages(msgs)
				if err != nil {
					m.cleanUpImport(communityID)
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					return
				}

				if m.DiscordImportMarkedAsCancelled(communityID) {
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- communityID
					return
				}

				// We're multiplying `chunksCount` by `0.25` so we leave 25% for additional save operations
				// 0.5 are the previous 50% of progress
				currentCount := ii + 1
				progressValue := calculateProgress(i+1, totalImportChunkCount, 0.5+(float32(currentCount)/float32(chunksCount))*0.25)
				importProgress.UpdateTaskProgress(discord.ImportMessagesTask, progressValue)
				progressUpdates <- importProgress

				// We slow down the saving of message chunks to keep the database responsive
				if currentCount < chunksCount {
					time.Sleep(2 * time.Second)
				}
			}

			// Get slice of all values in `messagesToSave` map

			var messages = make([]*common.Message, 0, len(messagesToSave))
			for _, msg := range messagesToSave {
				messages = append(messages, msg)
			}

			// Same as above, we save these messages in chunks so we don't block
			// the database for a longer period of time
			messageChunks := chunkSlice(messages, maxChunkSizeMessages)
			chunksCount = len(messageChunks)

			for ii, msgs := range messageChunks {
				m.communitiesManager.LogStdout(fmt.Sprintf("saving %d/%d chunk with %d app messages", ii+1, chunksCount, len(msgs)))
				err = m.persistence.SaveMessages(msgs)
				if err != nil {
					m.cleanUpImport(communityID)
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					return
				}

				if m.DiscordImportMarkedAsCancelled(communityID) {
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- communityID
					return
				}

				// 0.75 are the previous 75% of progress, hence we multiply our chunk progress
				// by 0.25
				currentCount := ii + 1
				progressValue := calculateProgress(i+1, totalImportChunkCount, 0.75+(float32(currentCount)/float32(chunksCount))*0.25)
				// progressValue := 0.75 + ((float32(currentCount) / float32(chunksCount)) * 0.25)
				importProgress.UpdateTaskProgress(discord.ImportMessagesTask, progressValue)
				progressUpdates <- importProgress

				// We slow down the saving of message chunks to keep the database responsive
				if currentCount < chunksCount {
					time.Sleep(2 * time.Second)
				}
			}

			pinMessageChunks := chunkSlice(pinMessagesToSave, maxChunkSizeMessages)
			for _, pinMsgs := range pinMessageChunks {
				err = m.persistence.SavePinMessages(pinMsgs)
				if err != nil {
					m.cleanUpImport(communityID)
					importProgress.AddTaskError(discord.ImportMessagesTask, discord.Error(err.Error()))
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					return
				}

				if m.DiscordImportMarkedAsCancelled(communityID) {
					importProgress.StopTask(discord.ImportMessagesTask)
					progressUpdates <- importProgress
					cancel <- communityID
					return
				}
			}

			totalAssetsCount := len(messageAttachmentsToDownload) + len(authorProfilesToSave)
			var assetCounter discord.AssetCounter

			var wg sync.WaitGroup

			for id, author := range authorProfilesToSave {
				wg.Add(1)
				go func(id string, author *protobuf.DiscordMessageAuthor) {
					defer wg.Done()

					m.communitiesManager.LogStdout(fmt.Sprintf("downloading asset %d/%d", assetCounter.Value()+1, totalAssetsCount))
					imagePayload, err := discord.DownloadAvatarAsset(author.AvatarUrl)
					if err != nil {
						errmsg := fmt.Sprintf("Couldn't download profile avatar '%s': %s", author.AvatarUrl, err.Error())
						importProgress.AddTaskError(
							discord.DownloadAssetsTask,
							discord.Warning(errmsg),
						)
						progressUpdates <- importProgress
						return
					}

					err = m.persistence.UpdateDiscordMessageAuthorImage(author.Id, imagePayload)
					if err != nil {
						importProgress.AddTaskError(discord.DownloadAssetsTask, discord.Warning(err.Error()))
						progressUpdates <- importProgress
						return
					}

					author.AvatarImagePayload = imagePayload
					authorProfilesToSave[id] = author

					if m.DiscordImportMarkedAsCancelled(discordCommunity.IDString()) {
						importProgress.StopTask(discord.DownloadAssetsTask)
						progressUpdates <- importProgress
						cancel <- discordCommunity.IDString()
						return
					}

					assetCounter.Increase()
					progressValue := calculateProgress(i+1, totalImportChunkCount, (float32(assetCounter.Value())/float32(totalAssetsCount))*0.5)
					importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
					progressUpdates <- importProgress

				}(id, author)
			}
			wg.Wait()

			if m.DiscordImportMarkedAsCancelled(communityID) {
				importProgress.StopTask(discord.DownloadAssetsTask)
				progressUpdates <- importProgress
				cancel <- communityID
				return
			}

			for idxRange := range gopart.Partition(len(messageAttachmentsToDownload), 100) {
				attachments := messageAttachmentsToDownload[idxRange.Low:idxRange.High]
				wg.Add(1)
				go func(attachments []*protobuf.DiscordMessageAttachment) {
					defer wg.Done()
					for ii, attachment := range attachments {

						m.communitiesManager.LogStdout(fmt.Sprintf("downloading asset %d/%d", assetCounter.Value()+1, totalAssetsCount))

						assetPayload, contentType, err := discord.DownloadAsset(attachment.Url)
						if err != nil {
							errmsg := fmt.Sprintf("Couldn't download message attachment '%s': %s", attachment.Url, err.Error())
							importProgress.AddTaskError(
								discord.DownloadAssetsTask,
								discord.Warning(errmsg),
							)
							progressUpdates <- importProgress
							continue
						}

						attachment.Payload = assetPayload
						attachment.ContentType = contentType
						messageAttachmentsToDownload[ii] = attachment

						if m.DiscordImportMarkedAsCancelled(communityID) {
							importProgress.StopTask(discord.DownloadAssetsTask)
							progressUpdates <- importProgress
							cancel <- communityID
							return
						}

						assetCounter.Increase()
						progressValue := calculateProgress(i+1, totalImportChunkCount, (float32(assetCounter.Value())/float32(totalAssetsCount))*0.5)
						importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
						progressUpdates <- importProgress
					}
				}(attachments)
			}
			wg.Wait()

			if m.DiscordImportMarkedAsCancelled(communityID) {
				importProgress.StopTask(discord.DownloadAssetsTask)
				progressUpdates <- importProgress
				cancel <- communityID
				return
			}

			attachmentChunks := chunkAttachmentsByByteSize(messageAttachmentsToDownload, maxChunkSizeBytes)
			chunksCount = len(attachmentChunks)

			for ii, attachments := range attachmentChunks {
				m.communitiesManager.LogStdout(fmt.Sprintf("saving %d/%d chunk with %d discord message attachments", ii+1, chunksCount, len(attachments)))
				err = m.persistence.SaveDiscordMessageAttachments(attachments)
				if err != nil {
					m.cleanUpImport(communityID)
					importProgress.AddTaskError(discord.DownloadAssetsTask, discord.Error(err.Error()))
					importProgress.Stop()
					progressUpdates <- importProgress
					return
				}

				if m.DiscordImportMarkedAsCancelled(communityID) {
					importProgress.StopTask(discord.DownloadAssetsTask)
					progressUpdates <- importProgress
					cancel <- communityID
					return
				}

				// 0.5 are the previous 50% of progress, hence we multiply our chunk progress
				// by 0.5
				currentCount := ii + 1
				progressValue := calculateProgress(i+1, totalImportChunkCount, 0.5+(float32(currentCount)/float32(chunksCount))*0.5)
				importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
				progressUpdates <- importProgress

				// We slow down the saving of attachment chunks to keep the database responsive
				if currentCount < chunksCount {
					time.Sleep(2 * time.Second)
				}
			}

			if len(attachmentChunks) == 0 {
				progressValue := calculateProgress(i+1, totalImportChunkCount, 1.0)
				importProgress.UpdateTaskProgress(discord.DownloadAssetsTask, progressValue)
			}

			_, err := m.transport.JoinPublic(processedChannelIds[channel.Channel.ID])
			if err != nil {
				m.logger.Error("failed to load filter for chat", zap.Error(err))
				continue
			}

			wakuChatMessages, err := m.chatMessagesToWakuMessages(messages, discordCommunity)
			if err != nil {
				m.logger.Error("failed to convert chat messages into waku messages", zap.Error(err))
				continue
			}

			wakuPinMessages, err := m.pinMessagesToWakuMessages(pinMessagesToSave, discordCommunity)
			if err != nil {
				m.logger.Error("failed to convert pin messages into waku messages", zap.Error(err))
				continue
			}

			wakuMessages := append(wakuChatMessages, wakuPinMessages...)

			topics, err := m.communitiesManager.GetCommunityChatsTopics(discordCommunity.ID())
			if err != nil {
				m.logger.Error("failed to get community chat topics", zap.Error(err))
				continue
			}

			startDate := time.Unix(int64(exportData.OldestMessageTimestamp), 0)
			endDate := time.Now()

			_, err = m.communitiesManager.CreateHistoryArchiveTorrentFromMessages(
				discordCommunity.ID(),
				wakuMessages,
				topics,
				startDate,
				endDate,
				messageArchiveInterval,
				discordCommunity.Encrypted(),
			)
			if err != nil {
				m.logger.Error("failed to create history archive torrent", zap.Error(err))
				continue
			}

			if m.torrentClientReady() && communitySettings.HistoryArchiveSupportEnabled {

				err = m.communitiesManager.SeedHistoryArchiveTorrent(discordCommunity.ID())
				if err != nil {
					m.logger.Error("failed to seed history archive", zap.Error(err))
				}
				go m.communitiesManager.StartHistoryArchiveTasksInterval(discordCommunity, messageArchiveInterval)
			}
		}

		err = m.publishOrg(discordCommunity, false)
		if err != nil {
			m.cleanUpImport(communityID)
			importProgress.AddTaskError(discord.InitCommunityTask, discord.Error(err.Error()))
			importProgress.Stop()
			progressUpdates <- importProgress
			return
		}

		if m.DiscordImportMarkedAsCancelled(communityID) {
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			cancel <- communityID
			return
		}

		// Chats need to be saved after the community has been published,
		// hence we make this part of the `InitCommunityTask`
		err = m.saveChats(chatsToSave)
		if err != nil {
			m.cleanUpImport(communityID)
			importProgress.AddTaskError(discord.InitCommunityTask, discord.Error(err.Error()))
			importProgress.Stop()
			progressUpdates <- importProgress
			return
		}

		importProgress.UpdateTaskProgress(discord.InitCommunityTask, 0.15)
		progressUpdates <- importProgress

		if m.DiscordImportMarkedAsCancelled(communityID) {
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			cancel <- communityID
			return
		}

		// Init the community filter so we can receive messages on the community
		_, err = m.InitCommunityFilters([]transport.CommunityFilterToInitialize{{
			Shard:   discordCommunity.Shard(),
			PrivKey: discordCommunity.PrivateKey(),
		}})
		if err != nil {
			m.cleanUpImport(communityID)
			importProgress.AddTaskError(discord.InitCommunityTask, discord.Error(err.Error()))
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			return
		}
		importProgress.UpdateTaskProgress(discord.InitCommunityTask, 0.25)
		progressUpdates <- importProgress

		if m.DiscordImportMarkedAsCancelled(communityID) {
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			cancel <- communityID
			return
		}

		_, err = m.transport.InitPublicFilters(m.DefaultFilters(discordCommunity))
		if err != nil {
			m.cleanUpImport(communityID)
			importProgress.AddTaskError(discord.InitCommunityTask, discord.Error(err.Error()))
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			return
		}

		importProgress.UpdateTaskProgress(discord.InitCommunityTask, 0.5)
		progressUpdates <- importProgress

		if m.DiscordImportMarkedAsCancelled(communityID) {
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			cancel <- communityID
			return
		}

		filters := m.transport.Filters()
		_, err = m.scheduleSyncFilters(filters)
		if err != nil {
			m.cleanUpImport(communityID)
			importProgress.AddTaskError(discord.InitCommunityTask, discord.Error(err.Error()))
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			return
		}
		importProgress.UpdateTaskProgress(discord.InitCommunityTask, 0.75)
		progressUpdates <- importProgress

		if m.DiscordImportMarkedAsCancelled(communityID) {
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			cancel <- communityID
			return
		}

		err = m.reregisterForPushNotifications()
		if err != nil {
			m.cleanUpImport(communityID)
			importProgress.AddTaskError(discord.InitCommunityTask, discord.Error(err.Error()))
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			return
		}
		importProgress.UpdateTaskProgress(discord.InitCommunityTask, 1)
		progressUpdates <- importProgress

		if m.DiscordImportMarkedAsCancelled(communityID) {
			importProgress.StopTask(discord.InitCommunityTask)
			progressUpdates <- importProgress
			cancel <- communityID
			return
		}

		m.config.messengerSignalsHandler.DiscordCommunityImportFinished(communityID)
		close(done)
	}()
}
