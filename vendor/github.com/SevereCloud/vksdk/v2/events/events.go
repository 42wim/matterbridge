/*
Package events for community events handling.

See more https://vk.com/dev/groups_events
*/
package events // import "github.com/SevereCloud/vksdk/v2/events"

import (
	"context"
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/internal"
)

// EventType type.
type EventType string

// EventType list.
const (
	EventConfirmation                  = "confirmation"
	EventMessageNew                    = "message_new"
	EventMessageReply                  = "message_reply"
	EventMessageEdit                   = "message_edit"
	EventMessageAllow                  = "message_allow"
	EventMessageDeny                   = "message_deny"
	EventMessageTypingState            = "message_typing_state"
	EventMessageEvent                  = "message_event"
	EventPhotoNew                      = "photo_new"
	EventPhotoCommentNew               = "photo_comment_new"
	EventPhotoCommentEdit              = "photo_comment_edit"
	EventPhotoCommentRestore           = "photo_comment_restore"
	EventPhotoCommentDelete            = "photo_comment_delete"
	EventAudioNew                      = "audio_new"
	EventVideoNew                      = "video_new"
	EventVideoCommentNew               = "video_comment_new"
	EventVideoCommentEdit              = "video_comment_edit"
	EventVideoCommentRestore           = "video_comment_restore"
	EventVideoCommentDelete            = "video_comment_delete"
	EventWallPostNew                   = "wall_post_new"
	EventWallRepost                    = "wall_repost"
	EventWallReplyNew                  = "wall_reply_new"
	EventWallReplyEdit                 = "wall_reply_edit"
	EventWallReplyRestore              = "wall_reply_restore"
	EventWallReplyDelete               = "wall_reply_delete"
	EventBoardPostNew                  = "board_post_new"
	EventBoardPostEdit                 = "board_post_edit"
	EventBoardPostRestore              = "board_post_restore"
	EventBoardPostDelete               = "board_post_delete"
	EventMarketCommentNew              = "market_comment_new"
	EventMarketCommentEdit             = "market_comment_edit"
	EventMarketCommentRestore          = "market_comment_restore"
	EventMarketCommentDelete           = "market_comment_delete"
	EventMarketOrderNew                = "market_order_new"
	EventMarketOrderEdit               = "market_order_edit"
	EventGroupLeave                    = "group_leave"
	EventGroupJoin                     = "group_join"
	EventUserBlock                     = "user_block"
	EventUserUnblock                   = "user_unblock"
	EventPollVoteNew                   = "poll_vote_new"
	EventGroupOfficersEdit             = "group_officers_edit"
	EventGroupChangeSettings           = "group_change_settings"
	EventGroupChangePhoto              = "group_change_photo"
	EventVkpayTransaction              = "vkpay_transaction"
	EventLeadFormsNew                  = "lead_forms_new"
	EventAppPayload                    = "app_payload"
	EventMessageRead                   = "message_read"
	EventLikeAdd                       = "like_add"
	EventLikeRemove                    = "like_remove"
	EventDonutSubscriptionCreate       = "donut_subscription_create"
	EventDonutSubscriptionProlonged    = "donut_subscription_prolonged"
	EventDonutSubscriptionExpired      = "donut_subscription_expired"
	EventDonutSubscriptionCancelled    = "donut_subscription_cancelled"
	EventDonutSubscriptionPriceChanged = "donut_subscription_price_changed"
	EventDonutMoneyWithdraw            = "donut_money_withdraw"
	EventDonutMoneyWithdrawError       = "donut_money_withdraw_error"
)

// GroupEvent struct.
type GroupEvent struct {
	Type    EventType       `json:"type"`
	Object  json.RawMessage `json:"object"`
	GroupID int             `json:"group_id"`
	EventID string          `json:"event_id"`
	V       string          `json:"v"`
	Secret  string          `json:"secret"`
}

// FuncList struct.
type FuncList struct {
	messageNew                    []func(context.Context, MessageNewObject)
	messageReply                  []func(context.Context, MessageReplyObject)
	messageEdit                   []func(context.Context, MessageEditObject)
	messageAllow                  []func(context.Context, MessageAllowObject)
	messageDeny                   []func(context.Context, MessageDenyObject)
	messageTypingState            []func(context.Context, MessageTypingStateObject)
	messageEvent                  []func(context.Context, MessageEventObject)
	photoNew                      []func(context.Context, PhotoNewObject)
	photoCommentNew               []func(context.Context, PhotoCommentNewObject)
	photoCommentEdit              []func(context.Context, PhotoCommentEditObject)
	photoCommentRestore           []func(context.Context, PhotoCommentRestoreObject)
	photoCommentDelete            []func(context.Context, PhotoCommentDeleteObject)
	audioNew                      []func(context.Context, AudioNewObject)
	videoNew                      []func(context.Context, VideoNewObject)
	videoCommentNew               []func(context.Context, VideoCommentNewObject)
	videoCommentEdit              []func(context.Context, VideoCommentEditObject)
	videoCommentRestore           []func(context.Context, VideoCommentRestoreObject)
	videoCommentDelete            []func(context.Context, VideoCommentDeleteObject)
	wallPostNew                   []func(context.Context, WallPostNewObject)
	wallRepost                    []func(context.Context, WallRepostObject)
	wallReplyNew                  []func(context.Context, WallReplyNewObject)
	wallReplyEdit                 []func(context.Context, WallReplyEditObject)
	wallReplyRestore              []func(context.Context, WallReplyRestoreObject)
	wallReplyDelete               []func(context.Context, WallReplyDeleteObject)
	boardPostNew                  []func(context.Context, BoardPostNewObject)
	boardPostEdit                 []func(context.Context, BoardPostEditObject)
	boardPostRestore              []func(context.Context, BoardPostRestoreObject)
	boardPostDelete               []func(context.Context, BoardPostDeleteObject)
	marketCommentNew              []func(context.Context, MarketCommentNewObject)
	marketCommentEdit             []func(context.Context, MarketCommentEditObject)
	marketCommentRestore          []func(context.Context, MarketCommentRestoreObject)
	marketCommentDelete           []func(context.Context, MarketCommentDeleteObject)
	marketOrderNew                []func(context.Context, MarketOrderNewObject)
	marketOrderEdit               []func(context.Context, MarketOrderEditObject)
	groupLeave                    []func(context.Context, GroupLeaveObject)
	groupJoin                     []func(context.Context, GroupJoinObject)
	userBlock                     []func(context.Context, UserBlockObject)
	userUnblock                   []func(context.Context, UserUnblockObject)
	pollVoteNew                   []func(context.Context, PollVoteNewObject)
	groupOfficersEdit             []func(context.Context, GroupOfficersEditObject)
	groupChangeSettings           []func(context.Context, GroupChangeSettingsObject)
	groupChangePhoto              []func(context.Context, GroupChangePhotoObject)
	vkpayTransaction              []func(context.Context, VkpayTransactionObject)
	leadFormsNew                  []func(context.Context, LeadFormsNewObject)
	appPayload                    []func(context.Context, AppPayloadObject)
	messageRead                   []func(context.Context, MessageReadObject)
	likeAdd                       []func(context.Context, LikeAddObject)
	likeRemove                    []func(context.Context, LikeRemoveObject)
	donutSubscriptionCreate       []func(context.Context, DonutSubscriptionCreateObject)
	donutSubscriptionProlonged    []func(context.Context, DonutSubscriptionProlongedObject)
	donutSubscriptionExpired      []func(context.Context, DonutSubscriptionExpiredObject)
	donutSubscriptionCancelled    []func(context.Context, DonutSubscriptionCancelledObject)
	donutSubscriptionPriceChanged []func(context.Context, DonutSubscriptionPriceChangedObject)
	donutMoneyWithdraw            []func(context.Context, DonutMoneyWithdrawObject)
	donutMoneyWithdrawError       []func(context.Context, DonutMoneyWithdrawErrorObject)
	special                       map[EventType][]func(context.Context, GroupEvent)
	eventsList                    []EventType

	goroutine bool
}

// NewFuncList returns a new FuncList.
func NewFuncList() *FuncList {
	return &FuncList{
		special: make(map[EventType][]func(context.Context, GroupEvent)),
	}
}

// Handler group event handler.
func (fl FuncList) Handler(ctx context.Context, e GroupEvent) error { //nolint:gocyclo
	ctx = context.WithValue(ctx, internal.GroupIDKey, e.GroupID)
	ctx = context.WithValue(ctx, internal.EventIDKey, e.EventID)
	ctx = context.WithValue(ctx, internal.EventVersionKey, e.V)

	if sliceFunc, ok := fl.special[e.Type]; ok {
		for _, f := range sliceFunc {
			f := f

			if fl.goroutine {
				go func() { f(ctx, e) }()
			} else {
				f(ctx, e)
			}
		}
	}

	switch e.Type {
	case EventMessageNew:
		var obj MessageNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMessageReply:
		var obj MessageReplyObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageReply {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMessageEdit:
		var obj MessageEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMessageAllow:
		var obj MessageAllowObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageAllow {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMessageDeny:
		var obj MessageDenyObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageDeny {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMessageTypingState: // На основе ответа
		var obj MessageTypingStateObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageTypingState {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMessageEvent:
		var obj MessageEventObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageEvent {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventPhotoNew:
		var obj PhotoNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.photoNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventPhotoCommentNew:
		var obj PhotoCommentNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.photoCommentNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventPhotoCommentEdit:
		var obj PhotoCommentEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.photoCommentEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventPhotoCommentRestore:
		var obj PhotoCommentRestoreObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.photoCommentRestore {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventPhotoCommentDelete:
		var obj PhotoCommentDeleteObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.photoCommentDelete {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventAudioNew:
		var obj AudioNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.audioNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventVideoNew:
		var obj VideoNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.videoNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventVideoCommentNew:
		var obj VideoCommentNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.videoCommentNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventVideoCommentEdit:
		var obj VideoCommentEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.videoCommentEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventVideoCommentRestore:
		var obj VideoCommentRestoreObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.videoCommentRestore {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventVideoCommentDelete:
		var obj VideoCommentDeleteObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.videoCommentDelete {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventWallPostNew:
		var obj WallPostNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.wallPostNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventWallRepost:
		var obj WallRepostObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.wallRepost {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventWallReplyNew:
		var obj WallReplyNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.wallReplyNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventWallReplyEdit:
		var obj WallReplyEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.wallReplyEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventWallReplyRestore:
		var obj WallReplyRestoreObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.wallReplyRestore {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventWallReplyDelete:
		var obj WallReplyDeleteObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.wallReplyDelete {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventBoardPostNew:
		var obj BoardPostNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.boardPostNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventBoardPostEdit:
		var obj BoardPostEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.boardPostEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventBoardPostRestore:
		var obj BoardPostRestoreObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.boardPostRestore {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventBoardPostDelete:
		var obj BoardPostDeleteObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.boardPostDelete {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMarketCommentNew:
		var obj MarketCommentNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.marketCommentNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMarketCommentEdit:
		var obj MarketCommentEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.marketCommentEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMarketCommentRestore:
		var obj MarketCommentRestoreObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.marketCommentRestore {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMarketCommentDelete:
		var obj MarketCommentDeleteObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.marketCommentDelete {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMarketOrderNew:
		var obj MarketOrderNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.marketOrderNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMarketOrderEdit:
		var obj MarketOrderEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.marketOrderEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventGroupLeave:
		var obj GroupLeaveObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.groupLeave {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventGroupJoin:
		var obj GroupJoinObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.groupJoin {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventUserBlock:
		var obj UserBlockObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.userBlock {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventUserUnblock:
		var obj UserUnblockObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.userUnblock {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventPollVoteNew:
		var obj PollVoteNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.pollVoteNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventGroupOfficersEdit:
		var obj GroupOfficersEditObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.groupOfficersEdit {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventGroupChangeSettings:
		var obj GroupChangeSettingsObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.groupChangeSettings {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventGroupChangePhoto:
		var obj GroupChangePhotoObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.groupChangePhoto {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventVkpayTransaction:
		var obj VkpayTransactionObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.vkpayTransaction {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventLeadFormsNew:
		var obj LeadFormsNewObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.leadFormsNew {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventAppPayload:
		var obj AppPayloadObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.appPayload {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventMessageRead:
		var obj MessageReadObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.messageRead {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventLikeAdd:
		var obj LikeAddObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.likeAdd {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventLikeRemove:
		var obj LikeRemoveObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.likeRemove {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventDonutSubscriptionCreate:
		var obj DonutSubscriptionCreateObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.donutSubscriptionCreate {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventDonutSubscriptionProlonged:
		var obj DonutSubscriptionProlongedObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.donutSubscriptionProlonged {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventDonutSubscriptionExpired:
		var obj DonutSubscriptionExpiredObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.donutSubscriptionExpired {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventDonutSubscriptionCancelled:
		var obj DonutSubscriptionCancelledObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.donutSubscriptionCancelled {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventDonutSubscriptionPriceChanged:
		var obj DonutSubscriptionPriceChangedObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.donutSubscriptionPriceChanged {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventDonutMoneyWithdraw:
		var obj DonutMoneyWithdrawObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.donutMoneyWithdraw {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	case EventDonutMoneyWithdrawError:
		var obj DonutMoneyWithdrawErrorObject
		if err := json.Unmarshal(e.Object, &obj); err != nil {
			return err
		}

		for _, f := range fl.donutMoneyWithdrawError {
			f := f

			if fl.goroutine {
				go func() { f(ctx, obj) }()
			} else {
				f(ctx, obj)
			}
		}
	}

	return nil
}

// ListEvents return list of events.
func (fl FuncList) ListEvents() []EventType {
	return fl.eventsList
}

// Goroutine invoke functions in a goroutine.
func (fl *FuncList) Goroutine(v bool) {
	fl.goroutine = v
}

// OnEvent handler.
func (fl *FuncList) OnEvent(eventType EventType, f func(context.Context, GroupEvent)) {
	if fl.special == nil {
		fl.special = make(map[EventType][]func(context.Context, GroupEvent))
	}

	fl.special[eventType] = append(fl.special[eventType], f)
	fl.eventsList = append(fl.eventsList, eventType)
}

// MessageNew handler.
func (fl *FuncList) MessageNew(f func(context.Context, MessageNewObject)) {
	fl.messageNew = append(fl.messageNew, f)
	fl.eventsList = append(fl.eventsList, EventMessageNew)
}

// MessageReply handler.
func (fl *FuncList) MessageReply(f func(context.Context, MessageReplyObject)) {
	fl.messageReply = append(fl.messageReply, f)
	fl.eventsList = append(fl.eventsList, EventMessageReply)
}

// MessageEdit handler.
func (fl *FuncList) MessageEdit(f func(context.Context, MessageEditObject)) {
	fl.messageEdit = append(fl.messageEdit, f)
	fl.eventsList = append(fl.eventsList, EventMessageEdit)
}

// MessageAllow handler.
func (fl *FuncList) MessageAllow(f func(context.Context, MessageAllowObject)) {
	fl.messageAllow = append(fl.messageAllow, f)
	fl.eventsList = append(fl.eventsList, EventMessageAllow)
}

// MessageDeny handler.
func (fl *FuncList) MessageDeny(f func(context.Context, MessageDenyObject)) {
	fl.messageDeny = append(fl.messageDeny, f)
	fl.eventsList = append(fl.eventsList, EventMessageDeny)
}

// MessageTypingState handler.
func (fl *FuncList) MessageTypingState(f func(context.Context, MessageTypingStateObject)) {
	fl.messageTypingState = append(fl.messageTypingState, f)
	fl.eventsList = append(fl.eventsList, EventMessageTypingState)
}

// MessageEvent handler.
func (fl *FuncList) MessageEvent(f func(context.Context, MessageEventObject)) {
	fl.messageEvent = append(fl.messageEvent, f)
	fl.eventsList = append(fl.eventsList, EventMessageEvent)
}

// PhotoNew handler.
func (fl *FuncList) PhotoNew(f func(context.Context, PhotoNewObject)) {
	fl.photoNew = append(fl.photoNew, f)
	fl.eventsList = append(fl.eventsList, EventPhotoNew)
}

// PhotoCommentNew handler.
func (fl *FuncList) PhotoCommentNew(f func(context.Context, PhotoCommentNewObject)) {
	fl.photoCommentNew = append(fl.photoCommentNew, f)
	fl.eventsList = append(fl.eventsList, EventPhotoCommentNew)
}

// PhotoCommentEdit handler.
func (fl *FuncList) PhotoCommentEdit(f func(context.Context, PhotoCommentEditObject)) {
	fl.photoCommentEdit = append(fl.photoCommentEdit, f)
	fl.eventsList = append(fl.eventsList, EventPhotoCommentEdit)
}

// PhotoCommentRestore handler.
func (fl *FuncList) PhotoCommentRestore(f func(context.Context, PhotoCommentRestoreObject)) {
	fl.photoCommentRestore = append(fl.photoCommentRestore, f)
	fl.eventsList = append(fl.eventsList, EventPhotoCommentRestore)
}

// PhotoCommentDelete handler.
func (fl *FuncList) PhotoCommentDelete(f func(context.Context, PhotoCommentDeleteObject)) {
	fl.photoCommentDelete = append(fl.photoCommentDelete, f)
	fl.eventsList = append(fl.eventsList, EventPhotoCommentDelete)
}

// AudioNew handler.
func (fl *FuncList) AudioNew(f func(context.Context, AudioNewObject)) {
	fl.audioNew = append(fl.audioNew, f)
	fl.eventsList = append(fl.eventsList, EventAudioNew)
}

// VideoNew handler.
func (fl *FuncList) VideoNew(f func(context.Context, VideoNewObject)) {
	fl.videoNew = append(fl.videoNew, f)
	fl.eventsList = append(fl.eventsList, EventVideoNew)
}

// VideoCommentNew handler.
func (fl *FuncList) VideoCommentNew(f func(context.Context, VideoCommentNewObject)) {
	fl.videoCommentNew = append(fl.videoCommentNew, f)
	fl.eventsList = append(fl.eventsList, EventVideoCommentNew)
}

// VideoCommentEdit handler.
func (fl *FuncList) VideoCommentEdit(f func(context.Context, VideoCommentEditObject)) {
	fl.videoCommentEdit = append(fl.videoCommentEdit, f)
	fl.eventsList = append(fl.eventsList, EventVideoCommentEdit)
}

// VideoCommentRestore handler.
func (fl *FuncList) VideoCommentRestore(f func(context.Context, VideoCommentRestoreObject)) {
	fl.videoCommentRestore = append(fl.videoCommentRestore, f)
	fl.eventsList = append(fl.eventsList, EventVideoCommentRestore)
}

// VideoCommentDelete handler.
func (fl *FuncList) VideoCommentDelete(f func(context.Context, VideoCommentDeleteObject)) {
	fl.videoCommentDelete = append(fl.videoCommentDelete, f)
	fl.eventsList = append(fl.eventsList, EventVideoCommentDelete)
}

// WallPostNew handler.
func (fl *FuncList) WallPostNew(f func(context.Context, WallPostNewObject)) {
	fl.wallPostNew = append(fl.wallPostNew, f)
	fl.eventsList = append(fl.eventsList, EventWallPostNew)
}

// WallRepost handler.
func (fl *FuncList) WallRepost(f func(context.Context, WallRepostObject)) {
	fl.wallRepost = append(fl.wallRepost, f)
	fl.eventsList = append(fl.eventsList, EventWallRepost)
}

// WallReplyNew handler.
func (fl *FuncList) WallReplyNew(f func(context.Context, WallReplyNewObject)) {
	fl.wallReplyNew = append(fl.wallReplyNew, f)
	fl.eventsList = append(fl.eventsList, EventWallReplyNew)
}

// WallReplyEdit handler.
func (fl *FuncList) WallReplyEdit(f func(context.Context, WallReplyEditObject)) {
	fl.wallReplyEdit = append(fl.wallReplyEdit, f)
	fl.eventsList = append(fl.eventsList, EventWallReplyEdit)
}

// WallReplyRestore handler.
func (fl *FuncList) WallReplyRestore(f func(context.Context, WallReplyRestoreObject)) {
	fl.wallReplyRestore = append(fl.wallReplyRestore, f)
	fl.eventsList = append(fl.eventsList, EventWallReplyRestore)
}

// WallReplyDelete handler.
func (fl *FuncList) WallReplyDelete(f func(context.Context, WallReplyDeleteObject)) {
	fl.wallReplyDelete = append(fl.wallReplyDelete, f)
	fl.eventsList = append(fl.eventsList, EventWallReplyDelete)
}

// BoardPostNew handler.
func (fl *FuncList) BoardPostNew(f func(context.Context, BoardPostNewObject)) {
	fl.boardPostNew = append(fl.boardPostNew, f)
	fl.eventsList = append(fl.eventsList, EventBoardPostNew)
}

// BoardPostEdit handler.
func (fl *FuncList) BoardPostEdit(f func(context.Context, BoardPostEditObject)) {
	fl.boardPostEdit = append(fl.boardPostEdit, f)
	fl.eventsList = append(fl.eventsList, EventBoardPostEdit)
}

// BoardPostRestore handler.
func (fl *FuncList) BoardPostRestore(f func(context.Context, BoardPostRestoreObject)) {
	fl.boardPostRestore = append(fl.boardPostRestore, f)
	fl.eventsList = append(fl.eventsList, EventBoardPostRestore)
}

// BoardPostDelete handler.
func (fl *FuncList) BoardPostDelete(f func(context.Context, BoardPostDeleteObject)) {
	fl.boardPostDelete = append(fl.boardPostDelete, f)
	fl.eventsList = append(fl.eventsList, EventBoardPostDelete)
}

// MarketCommentNew handler.
func (fl *FuncList) MarketCommentNew(f func(context.Context, MarketCommentNewObject)) {
	fl.marketCommentNew = append(fl.marketCommentNew, f)
	fl.eventsList = append(fl.eventsList, EventMarketCommentNew)
}

// MarketCommentEdit handler.
func (fl *FuncList) MarketCommentEdit(f func(context.Context, MarketCommentEditObject)) {
	fl.marketCommentEdit = append(fl.marketCommentEdit, f)
	fl.eventsList = append(fl.eventsList, EventMarketCommentEdit)
}

// MarketCommentRestore handler.
func (fl *FuncList) MarketCommentRestore(f func(context.Context, MarketCommentRestoreObject)) {
	fl.marketCommentRestore = append(fl.marketCommentRestore, f)
	fl.eventsList = append(fl.eventsList, EventMarketCommentRestore)
}

// MarketCommentDelete handler.
func (fl *FuncList) MarketCommentDelete(f func(context.Context, MarketCommentDeleteObject)) {
	fl.marketCommentDelete = append(fl.marketCommentDelete, f)
	fl.eventsList = append(fl.eventsList, EventMarketCommentDelete)
}

// MarketOrderNew handler.
func (fl *FuncList) MarketOrderNew(f func(context.Context, MarketOrderNewObject)) {
	fl.marketOrderNew = append(fl.marketOrderNew, f)
	fl.eventsList = append(fl.eventsList, EventMarketOrderNew)
}

// MarketOrderEdit handler.
func (fl *FuncList) MarketOrderEdit(f func(context.Context, MarketOrderEditObject)) {
	fl.marketOrderEdit = append(fl.marketOrderEdit, f)
	fl.eventsList = append(fl.eventsList, EventMarketOrderEdit)
}

// GroupLeave handler.
func (fl *FuncList) GroupLeave(f func(context.Context, GroupLeaveObject)) {
	fl.groupLeave = append(fl.groupLeave, f)
	fl.eventsList = append(fl.eventsList, EventGroupLeave)
}

// GroupJoin handler.
func (fl *FuncList) GroupJoin(f func(context.Context, GroupJoinObject)) {
	fl.groupJoin = append(fl.groupJoin, f)
	fl.eventsList = append(fl.eventsList, EventGroupJoin)
}

// UserBlock handler.
func (fl *FuncList) UserBlock(f func(context.Context, UserBlockObject)) {
	fl.userBlock = append(fl.userBlock, f)
	fl.eventsList = append(fl.eventsList, EventUserBlock)
}

// UserUnblock handler.
func (fl *FuncList) UserUnblock(f func(context.Context, UserUnblockObject)) {
	fl.userUnblock = append(fl.userUnblock, f)
	fl.eventsList = append(fl.eventsList, EventUserUnblock)
}

// PollVoteNew handler.
func (fl *FuncList) PollVoteNew(f func(context.Context, PollVoteNewObject)) {
	fl.pollVoteNew = append(fl.pollVoteNew, f)
	fl.eventsList = append(fl.eventsList, EventPollVoteNew)
}

// GroupOfficersEdit handler.
func (fl *FuncList) GroupOfficersEdit(f func(context.Context, GroupOfficersEditObject)) {
	fl.groupOfficersEdit = append(fl.groupOfficersEdit, f)
	fl.eventsList = append(fl.eventsList, EventGroupOfficersEdit)
}

// GroupChangeSettings handler.
func (fl *FuncList) GroupChangeSettings(f func(context.Context, GroupChangeSettingsObject)) {
	fl.groupChangeSettings = append(fl.groupChangeSettings, f)
	fl.eventsList = append(fl.eventsList, EventGroupChangeSettings)
}

// GroupChangePhoto handler.
func (fl *FuncList) GroupChangePhoto(f func(context.Context, GroupChangePhotoObject)) {
	fl.groupChangePhoto = append(fl.groupChangePhoto, f)
	fl.eventsList = append(fl.eventsList, EventGroupChangePhoto)
}

// VkpayTransaction handler.
func (fl *FuncList) VkpayTransaction(f func(context.Context, VkpayTransactionObject)) {
	fl.vkpayTransaction = append(fl.vkpayTransaction, f)
	fl.eventsList = append(fl.eventsList, EventVkpayTransaction)
}

// LeadFormsNew handler.
func (fl *FuncList) LeadFormsNew(f func(context.Context, LeadFormsNewObject)) {
	fl.leadFormsNew = append(fl.leadFormsNew, f)
	fl.eventsList = append(fl.eventsList, EventLeadFormsNew)
}

// AppPayload handler.
func (fl *FuncList) AppPayload(f func(context.Context, AppPayloadObject)) {
	fl.appPayload = append(fl.appPayload, f)
	fl.eventsList = append(fl.eventsList, EventAppPayload)
}

// MessageRead handler.
func (fl *FuncList) MessageRead(f func(context.Context, MessageReadObject)) {
	fl.messageRead = append(fl.messageRead, f)
	fl.eventsList = append(fl.eventsList, EventMessageRead)
}

// LikeAdd handler.
func (fl *FuncList) LikeAdd(f func(context.Context, LikeAddObject)) {
	fl.likeAdd = append(fl.likeAdd, f)
	fl.eventsList = append(fl.eventsList, EventLikeAdd)
}

// LikeRemove handler.
func (fl *FuncList) LikeRemove(f func(context.Context, LikeRemoveObject)) {
	fl.likeRemove = append(fl.likeRemove, f)
	fl.eventsList = append(fl.eventsList, EventLikeRemove)
}

// DonutSubscriptionCreate handler.
func (fl *FuncList) DonutSubscriptionCreate(f func(context.Context, DonutSubscriptionCreateObject)) {
	fl.donutSubscriptionCreate = append(fl.donutSubscriptionCreate, f)
	fl.eventsList = append(fl.eventsList, EventDonutSubscriptionCreate)
}

// DonutSubscriptionProlonged handler.
func (fl *FuncList) DonutSubscriptionProlonged(f func(context.Context, DonutSubscriptionProlongedObject)) {
	fl.donutSubscriptionProlonged = append(fl.donutSubscriptionProlonged, f)
	fl.eventsList = append(fl.eventsList, EventDonutSubscriptionProlonged)
}

// DonutSubscriptionExpired handler.
func (fl *FuncList) DonutSubscriptionExpired(f func(context.Context, DonutSubscriptionExpiredObject)) {
	fl.donutSubscriptionExpired = append(fl.donutSubscriptionExpired, f)
	fl.eventsList = append(fl.eventsList, EventDonutSubscriptionExpired)
}

// DonutSubscriptionCancelled handler.
func (fl *FuncList) DonutSubscriptionCancelled(f func(context.Context, DonutSubscriptionCancelledObject)) {
	fl.donutSubscriptionCancelled = append(fl.donutSubscriptionCancelled, f)
	fl.eventsList = append(fl.eventsList, EventDonutSubscriptionCancelled)
}

// DonutSubscriptionPriceChanged handler.
func (fl *FuncList) DonutSubscriptionPriceChanged(f func(context.Context, DonutSubscriptionPriceChangedObject)) {
	fl.donutSubscriptionPriceChanged = append(fl.donutSubscriptionPriceChanged, f)
	fl.eventsList = append(fl.eventsList, EventDonutSubscriptionPriceChanged)
}

// DonutMoneyWithdraw handler.
func (fl *FuncList) DonutMoneyWithdraw(f func(context.Context, DonutMoneyWithdrawObject)) {
	fl.donutMoneyWithdraw = append(fl.donutMoneyWithdraw, f)
	fl.eventsList = append(fl.eventsList, EventDonutMoneyWithdraw)
}

// DonutMoneyWithdrawError handler.
func (fl *FuncList) DonutMoneyWithdrawError(f func(context.Context, DonutMoneyWithdrawErrorObject)) {
	fl.donutMoneyWithdrawError = append(fl.donutMoneyWithdrawError, f)
	fl.eventsList = append(fl.eventsList, EventDonutMoneyWithdrawError)
}
