package tgbotapi

import (
	"net/url"
)

// NewMessage creates a new Message.
//
// chatID is where to send it, text is the message text.
func NewMessage(chatID int64, text string) MessageConfig {
	return MessageConfig{
		BaseChat: BaseChat{
			ChatID:           chatID,
			ReplyToMessageID: 0,
		},
		Text:                  text,
		DisableWebPagePreview: false,
	}
}

// NewDeleteMessage creates a request to delete a message.
func NewDeleteMessage(chatID int64, messageID int) DeleteMessageConfig {
	return DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: messageID,
	}
}

// NewMessageToChannel creates a new Message that is sent to a channel
// by username.
//
// username is the username of the channel, text is the message text,
// and the username should be in the form of `@username`.
func NewMessageToChannel(username string, text string) MessageConfig {
	return MessageConfig{
		BaseChat: BaseChat{
			ChannelUsername: username,
		},
		Text: text,
	}
}

// NewForward creates a new forward.
//
// chatID is where to send it, fromChatID is the source chat,
// and messageID is the ID of the original message.
func NewForward(chatID int64, fromChatID int64, messageID int) ForwardConfig {
	return ForwardConfig{
		BaseChat:   BaseChat{ChatID: chatID},
		FromChatID: fromChatID,
		MessageID:  messageID,
	}
}

// NewCopyMessage creates a new copy message.
//
// chatID is where to send it, fromChatID is the source chat,
// and messageID is the ID of the original message.
func NewCopyMessage(chatID int64, fromChatID int64, messageID int) CopyMessageConfig {
	return CopyMessageConfig{
		BaseChat:   BaseChat{ChatID: chatID},
		FromChatID: fromChatID,
		MessageID:  messageID,
	}
}

// NewPhoto creates a new sendPhoto request.
//
// chatID is where to send it, file is a string path to the file,
// FileReader, or FileBytes.
//
// Note that you must send animated GIFs as a document.
func NewPhoto(chatID int64, file RequestFileData) PhotoConfig {
	return PhotoConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
	}
}

// NewPhotoToChannel creates a new photo uploader to send a photo to a channel.
//
// Note that you must send animated GIFs as a document.
func NewPhotoToChannel(username string, file RequestFileData) PhotoConfig {
	return PhotoConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{
				ChannelUsername: username,
			},
			File: file,
		},
	}
}

// NewAudio creates a new sendAudio request.
func NewAudio(chatID int64, file RequestFileData) AudioConfig {
	return AudioConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
	}
}

// NewDocument creates a new sendDocument request.
func NewDocument(chatID int64, file RequestFileData) DocumentConfig {
	return DocumentConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
	}
}

// NewSticker creates a new sendSticker request.
func NewSticker(chatID int64, file RequestFileData) StickerConfig {
	return StickerConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
	}
}

// NewVideo creates a new sendVideo request.
func NewVideo(chatID int64, file RequestFileData) VideoConfig {
	return VideoConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
	}
}

// NewAnimation creates a new sendAnimation request.
func NewAnimation(chatID int64, file RequestFileData) AnimationConfig {
	return AnimationConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
	}
}

// NewVideoNote creates a new sendVideoNote request.
//
// chatID is where to send it, file is a string path to the file,
// FileReader, or FileBytes.
func NewVideoNote(chatID int64, length int, file RequestFileData) VideoNoteConfig {
	return VideoNoteConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
		Length: length,
	}
}

// NewVoice creates a new sendVoice request.
func NewVoice(chatID int64, file RequestFileData) VoiceConfig {
	return VoiceConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{ChatID: chatID},
			File:     file,
		},
	}
}

// NewMediaGroup creates a new media group. Files should be an array of
// two to ten InputMediaPhoto or InputMediaVideo.
func NewMediaGroup(chatID int64, files []interface{}) MediaGroupConfig {
	return MediaGroupConfig{
		ChatID: chatID,
		Media:  files,
	}
}

// NewInputMediaPhoto creates a new InputMediaPhoto.
func NewInputMediaPhoto(media RequestFileData) InputMediaPhoto {
	return InputMediaPhoto{
		BaseInputMedia{
			Type:  "photo",
			Media: media,
		},
	}
}

// NewInputMediaVideo creates a new InputMediaVideo.
func NewInputMediaVideo(media RequestFileData) InputMediaVideo {
	return InputMediaVideo{
		BaseInputMedia: BaseInputMedia{
			Type:  "video",
			Media: media,
		},
	}
}

// NewInputMediaAnimation creates a new InputMediaAnimation.
func NewInputMediaAnimation(media RequestFileData) InputMediaAnimation {
	return InputMediaAnimation{
		BaseInputMedia: BaseInputMedia{
			Type:  "animation",
			Media: media,
		},
	}
}

// NewInputMediaAudio creates a new InputMediaAudio.
func NewInputMediaAudio(media RequestFileData) InputMediaAudio {
	return InputMediaAudio{
		BaseInputMedia: BaseInputMedia{
			Type:  "audio",
			Media: media,
		},
	}
}

// NewInputMediaDocument creates a new InputMediaDocument.
func NewInputMediaDocument(media RequestFileData) InputMediaDocument {
	return InputMediaDocument{
		BaseInputMedia: BaseInputMedia{
			Type:  "document",
			Media: media,
		},
	}
}

// NewContact allows you to send a shared contact.
func NewContact(chatID int64, phoneNumber, firstName string) ContactConfig {
	return ContactConfig{
		BaseChat: BaseChat{
			ChatID: chatID,
		},
		PhoneNumber: phoneNumber,
		FirstName:   firstName,
	}
}

// NewLocation shares your location.
//
// chatID is where to send it, latitude and longitude are coordinates.
func NewLocation(chatID int64, latitude float64, longitude float64) LocationConfig {
	return LocationConfig{
		BaseChat: BaseChat{
			ChatID: chatID,
		},
		Latitude:  latitude,
		Longitude: longitude,
	}
}

// NewVenue allows you to send a venue and its location.
func NewVenue(chatID int64, title, address string, latitude, longitude float64) VenueConfig {
	return VenueConfig{
		BaseChat: BaseChat{
			ChatID: chatID,
		},
		Title:     title,
		Address:   address,
		Latitude:  latitude,
		Longitude: longitude,
	}
}

// NewChatAction sets a chat action.
// Actions last for 5 seconds, or until your next action.
//
// chatID is where to send it, action should be set via Chat constants.
func NewChatAction(chatID int64, action string) ChatActionConfig {
	return ChatActionConfig{
		BaseChat: BaseChat{ChatID: chatID},
		Action:   action,
	}
}

// NewUserProfilePhotos gets user profile photos.
//
// userID is the ID of the user you wish to get profile photos from.
func NewUserProfilePhotos(userID int64) UserProfilePhotosConfig {
	return UserProfilePhotosConfig{
		UserID: userID,
		Offset: 0,
		Limit:  0,
	}
}

// NewUpdate gets updates since the last Offset.
//
// offset is the last Update ID to include.
// You likely want to set this to the last Update ID plus 1.
func NewUpdate(offset int) UpdateConfig {
	return UpdateConfig{
		Offset:  offset,
		Limit:   0,
		Timeout: 0,
	}
}

// NewWebhook creates a new webhook.
//
// link is the url parsable link you wish to get the updates.
func NewWebhook(link string) (WebhookConfig, error) {
	u, err := url.Parse(link)

	if err != nil {
		return WebhookConfig{}, err
	}

	return WebhookConfig{
		URL: u,
	}, nil
}

// NewWebhookWithCert creates a new webhook with a certificate.
//
// link is the url you wish to get webhooks,
// file contains a string to a file, FileReader, or FileBytes.
func NewWebhookWithCert(link string, file RequestFileData) (WebhookConfig, error) {
	u, err := url.Parse(link)

	if err != nil {
		return WebhookConfig{}, err
	}

	return WebhookConfig{
		URL:         u,
		Certificate: file,
	}, nil
}

// NewInlineQueryResultArticle creates a new inline query article.
func NewInlineQueryResultArticle(id, title, messageText string) InlineQueryResultArticle {
	return InlineQueryResultArticle{
		Type:  "article",
		ID:    id,
		Title: title,
		InputMessageContent: InputTextMessageContent{
			Text: messageText,
		},
	}
}

// NewInlineQueryResultArticleMarkdown creates a new inline query article with Markdown parsing.
func NewInlineQueryResultArticleMarkdown(id, title, messageText string) InlineQueryResultArticle {
	return InlineQueryResultArticle{
		Type:  "article",
		ID:    id,
		Title: title,
		InputMessageContent: InputTextMessageContent{
			Text:      messageText,
			ParseMode: "Markdown",
		},
	}
}

// NewInlineQueryResultArticleMarkdownV2 creates a new inline query article with MarkdownV2 parsing.
func NewInlineQueryResultArticleMarkdownV2(id, title, messageText string) InlineQueryResultArticle {
	return InlineQueryResultArticle{
		Type:  "article",
		ID:    id,
		Title: title,
		InputMessageContent: InputTextMessageContent{
			Text:      messageText,
			ParseMode: "MarkdownV2",
		},
	}
}

// NewInlineQueryResultArticleHTML creates a new inline query article with HTML parsing.
func NewInlineQueryResultArticleHTML(id, title, messageText string) InlineQueryResultArticle {
	return InlineQueryResultArticle{
		Type:  "article",
		ID:    id,
		Title: title,
		InputMessageContent: InputTextMessageContent{
			Text:      messageText,
			ParseMode: "HTML",
		},
	}
}

// NewInlineQueryResultGIF creates a new inline query GIF.
func NewInlineQueryResultGIF(id, url string) InlineQueryResultGIF {
	return InlineQueryResultGIF{
		Type: "gif",
		ID:   id,
		URL:  url,
	}
}

// NewInlineQueryResultCachedGIF create a new inline query with cached photo.
func NewInlineQueryResultCachedGIF(id, gifID string) InlineQueryResultCachedGIF {
	return InlineQueryResultCachedGIF{
		Type:  "gif",
		ID:    id,
		GIFID: gifID,
	}
}

// NewInlineQueryResultMPEG4GIF creates a new inline query MPEG4 GIF.
func NewInlineQueryResultMPEG4GIF(id, url string) InlineQueryResultMPEG4GIF {
	return InlineQueryResultMPEG4GIF{
		Type: "mpeg4_gif",
		ID:   id,
		URL:  url,
	}
}

// NewInlineQueryResultCachedMPEG4GIF create a new inline query with cached MPEG4 GIF.
func NewInlineQueryResultCachedMPEG4GIF(id, MPEG4GIFID string) InlineQueryResultCachedMPEG4GIF {
	return InlineQueryResultCachedMPEG4GIF{
		Type:        "mpeg4_gif",
		ID:          id,
		MPEG4FileID: MPEG4GIFID,
	}
}

// NewInlineQueryResultPhoto creates a new inline query photo.
func NewInlineQueryResultPhoto(id, url string) InlineQueryResultPhoto {
	return InlineQueryResultPhoto{
		Type: "photo",
		ID:   id,
		URL:  url,
	}
}

// NewInlineQueryResultPhotoWithThumb creates a new inline query photo.
func NewInlineQueryResultPhotoWithThumb(id, url, thumb string) InlineQueryResultPhoto {
	return InlineQueryResultPhoto{
		Type:     "photo",
		ID:       id,
		URL:      url,
		ThumbURL: thumb,
	}
}

// NewInlineQueryResultCachedPhoto create a new inline query with cached photo.
func NewInlineQueryResultCachedPhoto(id, photoID string) InlineQueryResultCachedPhoto {
	return InlineQueryResultCachedPhoto{
		Type:    "photo",
		ID:      id,
		PhotoID: photoID,
	}
}

// NewInlineQueryResultVideo creates a new inline query video.
func NewInlineQueryResultVideo(id, url string) InlineQueryResultVideo {
	return InlineQueryResultVideo{
		Type: "video",
		ID:   id,
		URL:  url,
	}
}

// NewInlineQueryResultCachedVideo create a new inline query with cached video.
func NewInlineQueryResultCachedVideo(id, videoID, title string) InlineQueryResultCachedVideo {
	return InlineQueryResultCachedVideo{
		Type:    "video",
		ID:      id,
		VideoID: videoID,
		Title:   title,
	}
}

// NewInlineQueryResultCachedSticker create a new inline query with cached sticker.
func NewInlineQueryResultCachedSticker(id, stickerID, title string) InlineQueryResultCachedSticker {
	return InlineQueryResultCachedSticker{
		Type:      "sticker",
		ID:        id,
		StickerID: stickerID,
		Title:     title,
	}
}

// NewInlineQueryResultAudio creates a new inline query audio.
func NewInlineQueryResultAudio(id, url, title string) InlineQueryResultAudio {
	return InlineQueryResultAudio{
		Type:  "audio",
		ID:    id,
		URL:   url,
		Title: title,
	}
}

// NewInlineQueryResultCachedAudio create a new inline query with cached photo.
func NewInlineQueryResultCachedAudio(id, audioID string) InlineQueryResultCachedAudio {
	return InlineQueryResultCachedAudio{
		Type:    "audio",
		ID:      id,
		AudioID: audioID,
	}
}

// NewInlineQueryResultVoice creates a new inline query voice.
func NewInlineQueryResultVoice(id, url, title string) InlineQueryResultVoice {
	return InlineQueryResultVoice{
		Type:  "voice",
		ID:    id,
		URL:   url,
		Title: title,
	}
}

// NewInlineQueryResultCachedVoice create a new inline query with cached photo.
func NewInlineQueryResultCachedVoice(id, voiceID, title string) InlineQueryResultCachedVoice {
	return InlineQueryResultCachedVoice{
		Type:    "voice",
		ID:      id,
		VoiceID: voiceID,
		Title:   title,
	}
}

// NewInlineQueryResultDocument creates a new inline query document.
func NewInlineQueryResultDocument(id, url, title, mimeType string) InlineQueryResultDocument {
	return InlineQueryResultDocument{
		Type:     "document",
		ID:       id,
		URL:      url,
		Title:    title,
		MimeType: mimeType,
	}
}

// NewInlineQueryResultCachedDocument create a new inline query with cached photo.
func NewInlineQueryResultCachedDocument(id, documentID, title string) InlineQueryResultCachedDocument {
	return InlineQueryResultCachedDocument{
		Type:       "document",
		ID:         id,
		DocumentID: documentID,
		Title:      title,
	}
}

// NewInlineQueryResultLocation creates a new inline query location.
func NewInlineQueryResultLocation(id, title string, latitude, longitude float64) InlineQueryResultLocation {
	return InlineQueryResultLocation{
		Type:      "location",
		ID:        id,
		Title:     title,
		Latitude:  latitude,
		Longitude: longitude,
	}
}

// NewInlineQueryResultVenue creates a new inline query venue.
func NewInlineQueryResultVenue(id, title, address string, latitude, longitude float64) InlineQueryResultVenue {
	return InlineQueryResultVenue{
		Type:      "venue",
		ID:        id,
		Title:     title,
		Address:   address,
		Latitude:  latitude,
		Longitude: longitude,
	}
}

// NewEditMessageText allows you to edit the text of a message.
func NewEditMessageText(chatID int64, messageID int, text string) EditMessageTextConfig {
	return EditMessageTextConfig{
		BaseEdit: BaseEdit{
			ChatID:    chatID,
			MessageID: messageID,
		},
		Text: text,
	}
}

// NewEditMessageTextAndMarkup allows you to edit the text and replymarkup of a message.
func NewEditMessageTextAndMarkup(chatID int64, messageID int, text string, replyMarkup InlineKeyboardMarkup) EditMessageTextConfig {
	return EditMessageTextConfig{
		BaseEdit: BaseEdit{
			ChatID:      chatID,
			MessageID:   messageID,
			ReplyMarkup: &replyMarkup,
		},
		Text: text,
	}
}

// NewEditMessageCaption allows you to edit the caption of a message.
func NewEditMessageCaption(chatID int64, messageID int, caption string) EditMessageCaptionConfig {
	return EditMessageCaptionConfig{
		BaseEdit: BaseEdit{
			ChatID:    chatID,
			MessageID: messageID,
		},
		Caption: caption,
	}
}

// NewEditMessageReplyMarkup allows you to edit the inline
// keyboard markup.
func NewEditMessageReplyMarkup(chatID int64, messageID int, replyMarkup InlineKeyboardMarkup) EditMessageReplyMarkupConfig {
	return EditMessageReplyMarkupConfig{
		BaseEdit: BaseEdit{
			ChatID:      chatID,
			MessageID:   messageID,
			ReplyMarkup: &replyMarkup,
		},
	}
}

// NewRemoveKeyboard hides the keyboard, with the option for being selective
// or hiding for everyone.
func NewRemoveKeyboard(selective bool) ReplyKeyboardRemove {
	return ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      selective,
	}
}

// NewKeyboardButton creates a regular keyboard button.
func NewKeyboardButton(text string) KeyboardButton {
	return KeyboardButton{
		Text: text,
	}
}

// NewKeyboardButtonContact creates a keyboard button that requests
// user contact information upon click.
func NewKeyboardButtonContact(text string) KeyboardButton {
	return KeyboardButton{
		Text:           text,
		RequestContact: true,
	}
}

// NewKeyboardButtonLocation creates a keyboard button that requests
// user location information upon click.
func NewKeyboardButtonLocation(text string) KeyboardButton {
	return KeyboardButton{
		Text:            text,
		RequestLocation: true,
	}
}

// NewKeyboardButtonRow creates a row of keyboard buttons.
func NewKeyboardButtonRow(buttons ...KeyboardButton) []KeyboardButton {
	var row []KeyboardButton

	row = append(row, buttons...)

	return row
}

// NewReplyKeyboard creates a new regular keyboard with sane defaults.
func NewReplyKeyboard(rows ...[]KeyboardButton) ReplyKeyboardMarkup {
	var keyboard [][]KeyboardButton

	keyboard = append(keyboard, rows...)

	return ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard:       keyboard,
	}
}

// NewOneTimeReplyKeyboard creates a new one time keyboard.
func NewOneTimeReplyKeyboard(rows ...[]KeyboardButton) ReplyKeyboardMarkup {
	markup := NewReplyKeyboard(rows...)
	markup.OneTimeKeyboard = true
	return markup
}

// NewInlineKeyboardButtonData creates an inline keyboard button with text
// and data for a callback.
func NewInlineKeyboardButtonData(text, data string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:         text,
		CallbackData: &data,
	}
}

// NewInlineKeyboardButtonLoginURL creates an inline keyboard button with text
// which goes to a LoginURL.
func NewInlineKeyboardButtonLoginURL(text string, loginURL LoginURL) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:     text,
		LoginURL: &loginURL,
	}
}

// NewInlineKeyboardButtonURL creates an inline keyboard button with text
// which goes to a URL.
func NewInlineKeyboardButtonURL(text, url string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text: text,
		URL:  &url,
	}
}

// NewInlineKeyboardButtonSwitch creates an inline keyboard button with
// text which allows the user to switch to a chat or return to a chat.
func NewInlineKeyboardButtonSwitch(text, sw string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:              text,
		SwitchInlineQuery: &sw,
	}
}

// NewInlineKeyboardRow creates an inline keyboard row with buttons.
func NewInlineKeyboardRow(buttons ...InlineKeyboardButton) []InlineKeyboardButton {
	var row []InlineKeyboardButton

	row = append(row, buttons...)

	return row
}

// NewInlineKeyboardMarkup creates a new inline keyboard.
func NewInlineKeyboardMarkup(rows ...[]InlineKeyboardButton) InlineKeyboardMarkup {
	var keyboard [][]InlineKeyboardButton

	keyboard = append(keyboard, rows...)

	return InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

// NewCallback creates a new callback message.
func NewCallback(id, text string) CallbackConfig {
	return CallbackConfig{
		CallbackQueryID: id,
		Text:            text,
		ShowAlert:       false,
	}
}

// NewCallbackWithAlert creates a new callback message that alerts
// the user.
func NewCallbackWithAlert(id, text string) CallbackConfig {
	return CallbackConfig{
		CallbackQueryID: id,
		Text:            text,
		ShowAlert:       true,
	}
}

// NewInvoice creates a new Invoice request to the user.
func NewInvoice(chatID int64, title, description, payload, providerToken, startParameter, currency string, prices []LabeledPrice) InvoiceConfig {
	return InvoiceConfig{
		BaseChat:       BaseChat{ChatID: chatID},
		Title:          title,
		Description:    description,
		Payload:        payload,
		ProviderToken:  providerToken,
		StartParameter: startParameter,
		Currency:       currency,
		Prices:         prices}
}

// NewChatTitle allows you to update the title of a chat.
func NewChatTitle(chatID int64, title string) SetChatTitleConfig {
	return SetChatTitleConfig{
		ChatID: chatID,
		Title:  title,
	}
}

// NewChatDescription allows you to update the description of a chat.
func NewChatDescription(chatID int64, description string) SetChatDescriptionConfig {
	return SetChatDescriptionConfig{
		ChatID:      chatID,
		Description: description,
	}
}

// NewChatPhoto allows you to update the photo for a chat.
func NewChatPhoto(chatID int64, photo RequestFileData) SetChatPhotoConfig {
	return SetChatPhotoConfig{
		BaseFile: BaseFile{
			BaseChat: BaseChat{
				ChatID: chatID,
			},
			File: photo,
		},
	}
}

// NewDeleteChatPhoto allows you to delete the photo for a chat.
func NewDeleteChatPhoto(chatID int64) DeleteChatPhotoConfig {
	return DeleteChatPhotoConfig{
		ChatID: chatID,
	}
}

// NewPoll allows you to create a new poll.
func NewPoll(chatID int64, question string, options ...string) SendPollConfig {
	return SendPollConfig{
		BaseChat: BaseChat{
			ChatID: chatID,
		},
		Question:    question,
		Options:     options,
		IsAnonymous: true, // This is Telegram's default.
	}
}

// NewStopPoll allows you to stop a poll.
func NewStopPoll(chatID int64, messageID int) StopPollConfig {
	return StopPollConfig{
		BaseEdit{
			ChatID:    chatID,
			MessageID: messageID,
		},
	}
}

// NewDice allows you to send a random dice roll.
func NewDice(chatID int64) DiceConfig {
	return DiceConfig{
		BaseChat: BaseChat{
			ChatID: chatID,
		},
	}
}

// NewDiceWithEmoji allows you to send a random roll of one of many types.
//
// Emoji may be 🎲 (1-6), 🎯 (1-6), or 🏀 (1-5).
func NewDiceWithEmoji(chatID int64, emoji string) DiceConfig {
	return DiceConfig{
		BaseChat: BaseChat{
			ChatID: chatID,
		},
		Emoji: emoji,
	}
}

// NewBotCommandScopeDefault represents the default scope of bot commands.
func NewBotCommandScopeDefault() BotCommandScope {
	return BotCommandScope{Type: "default"}
}

// NewBotCommandScopeAllPrivateChats represents the scope of bot commands,
// covering all private chats.
func NewBotCommandScopeAllPrivateChats() BotCommandScope {
	return BotCommandScope{Type: "all_private_chats"}
}

// NewBotCommandScopeAllGroupChats represents the scope of bot commands,
// covering all group and supergroup chats.
func NewBotCommandScopeAllGroupChats() BotCommandScope {
	return BotCommandScope{Type: "all_group_chats"}
}

// NewBotCommandScopeAllChatAdministrators represents the scope of bot commands,
// covering all group and supergroup chat administrators.
func NewBotCommandScopeAllChatAdministrators() BotCommandScope {
	return BotCommandScope{Type: "all_chat_administrators"}
}

// NewBotCommandScopeChat represents the scope of bot commands, covering a
// specific chat.
func NewBotCommandScopeChat(chatID int64) BotCommandScope {
	return BotCommandScope{
		Type:   "chat",
		ChatID: chatID,
	}
}

// NewBotCommandScopeChatAdministrators represents the scope of bot commands,
// covering all administrators of a specific group or supergroup chat.
func NewBotCommandScopeChatAdministrators(chatID int64) BotCommandScope {
	return BotCommandScope{
		Type:   "chat_administrators",
		ChatID: chatID,
	}
}

// NewBotCommandScopeChatMember represents the scope of bot commands, covering a
// specific member of a group or supergroup chat.
func NewBotCommandScopeChatMember(chatID, userID int64) BotCommandScope {
	return BotCommandScope{
		Type:   "chat_member",
		ChatID: chatID,
		UserID: userID,
	}
}

// NewGetMyCommandsWithScope allows you to set the registered commands for a
// given scope.
func NewGetMyCommandsWithScope(scope BotCommandScope) GetMyCommandsConfig {
	return GetMyCommandsConfig{Scope: &scope}
}

// NewGetMyCommandsWithScopeAndLanguage allows you to set the registered
// commands for a given scope and language code.
func NewGetMyCommandsWithScopeAndLanguage(scope BotCommandScope, languageCode string) GetMyCommandsConfig {
	return GetMyCommandsConfig{Scope: &scope, LanguageCode: languageCode}
}

// NewSetMyCommands allows you to set the registered commands.
func NewSetMyCommands(commands ...BotCommand) SetMyCommandsConfig {
	return SetMyCommandsConfig{Commands: commands}
}

// NewSetMyCommandsWithScope allows you to set the registered commands for a given scope.
func NewSetMyCommandsWithScope(scope BotCommandScope, commands ...BotCommand) SetMyCommandsConfig {
	return SetMyCommandsConfig{Commands: commands, Scope: &scope}
}

// NewSetMyCommandsWithScopeAndLanguage allows you to set the registered commands for a given scope
// and language code.
func NewSetMyCommandsWithScopeAndLanguage(scope BotCommandScope, languageCode string, commands ...BotCommand) SetMyCommandsConfig {
	return SetMyCommandsConfig{Commands: commands, Scope: &scope, LanguageCode: languageCode}
}

// NewDeleteMyCommands allows you to delete the registered commands.
func NewDeleteMyCommands() DeleteMyCommandsConfig {
	return DeleteMyCommandsConfig{}
}

// NewDeleteMyCommandsWithScope allows you to delete the registered commands for a given
// scope.
func NewDeleteMyCommandsWithScope(scope BotCommandScope) DeleteMyCommandsConfig {
	return DeleteMyCommandsConfig{Scope: &scope}
}

// NewDeleteMyCommandsWithScopeAndLanguage allows you to delete the registered commands for a given
// scope and language code.
func NewDeleteMyCommandsWithScopeAndLanguage(scope BotCommandScope, languageCode string) DeleteMyCommandsConfig {
	return DeleteMyCommandsConfig{Scope: &scope, LanguageCode: languageCode}
}
