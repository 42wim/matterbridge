package api

import (
	"errors"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/object"
)

// ErrorType is the type of an error.
type ErrorType int

// Error returns the message of a ErrorType.
func (e ErrorType) Error() string {
	return fmt.Sprintf("api: error with code %d", e)
}

// Error codes. See https://vk.com/dev/errors
const (
	ErrNoType ErrorType = 0 // NoType error

	// Unknown error occurred
	//
	// Try again later.
	ErrUnknown ErrorType = 1

	// Application is disabled. Enable your application or use test mode
	//
	// You need to switch on the app in Settings
	// https://vk.com/editapp?id={Your API_ID}
	// or use the test mode (test_mode=1).
	ErrDisabled ErrorType = 2

	// Unknown method passed.
	//
	// Check the method name: http://vk.com/dev/methods
	ErrMethod    ErrorType = 3
	ErrSignature ErrorType = 4 // Incorrect signature

	// User authorization failed
	//
	// Make sure that you use a correct authorization type.
	ErrAuth ErrorType = 5

	// Too many requests per second
	//
	// Decrease the request frequency or use the execute method.
	// More details on frequency limits here:
	// https://vk.com/dev/api_requests
	ErrTooMany ErrorType = 6

	// Permission to perform this action is denied
	//
	// Make sure that your have received required permissions during the
	// authorization.
	// You can do it with the account.getAppPermissions method.
	// https://vk.com/dev/permissions
	ErrPermission ErrorType = 7

	// Invalid request
	//
	// Check the request syntax and used parameters list (it can be found on
	// a method description page).
	ErrRequest ErrorType = 8

	// Flood control
	//
	// You need to decrease the count of identical requests. For more efficient
	// work you may use execute.
	ErrFlood ErrorType = 9

	// Internal server error
	//
	// Try again later.
	ErrServer ErrorType = 10

	// In test mode application should be disabled or user should be authorized
	//
	// Switch the app off in Settings:
	//
	// 	https://vk.com/editapp?id={Your API_ID}
	//
	ErrEnabledInTest ErrorType = 11

	// Unable to compile code.
	ErrCompile ErrorType = 12

	// Runtime error occurred during code invocation.
	ErrRuntime ErrorType = 13

	// Captcha needed.
	//
	// See https://vk.com/dev/captcha_error
	ErrCaptcha ErrorType = 14

	// Access denied
	//
	// Make sure that you use correct identifiers and the content is available
	// for the user in the full version of the site.
	ErrAccess ErrorType = 15

	// HTTP authorization failed
	//
	// To avoid this error check if a user has the 'Use secure connection'
	// option enabled with the account.getInfo method.
	ErrAuthHTTPS ErrorType = 16

	// Validation required
	//
	// Make sure that you don't use a token received with
	// http://vk.com/dev/auth_mobile for a request from the server.
	// It's restricted.
	//
	// https://vk.com/dev/need_validation
	ErrAuthValidation ErrorType = 17
	ErrUserDeleted    ErrorType = 18 // User was deleted or banned
	ErrBlocked        ErrorType = 19 // Content blocked

	// Permission to perform this action is denied for non-standalone
	// applications.
	ErrMethodPermission ErrorType = 20

	// Permission to perform this action is allowed only for standalone and
	// OpenAPI applications.
	ErrMethodAds ErrorType = 21
	ErrUpload    ErrorType = 22 // Upload error

	// This method was disabled.
	//
	// All the methods available now are listed here: http://vk.com/dev/methods
	ErrMethodDisabled ErrorType = 23

	// Confirmation required
	//
	// In some cases VK requires to request action confirmation from the user
	// (for Standalone apps only).
	//
	// Following parameter is transmitted with the error message as well:
	//
	// confirmation_text – text of the message to be shown in the default
	// confirmation window.
	//
	// The app should display the default confirmation window
	// with text from confirmation_text and two buttons: "Continue" and
	// "Cancel".
	// If user confirms the action repeat the request with an extra parameter:
	//
	// 	confirm = 1.
	//
	// https://vk.com/dev/need_confirmation
	ErrNeedConfirmation      ErrorType = 24
	ErrNeedTokenConfirmation ErrorType = 25 // Token confirmation required
	ErrGroupAuth             ErrorType = 27 // Group authorization failed
	ErrAppAuth               ErrorType = 28 // Application authorization failed

	// Rate limit reached.
	//
	// More details on rate limits here: https://vk.com/dev/data_limits
	ErrRateLimit      ErrorType = 29
	ErrPrivateProfile ErrorType = 30 // This profile is private

	// Client version deprecated.
	ErrClientVersionDeprecated ErrorType = 34

	// Method execution was interrupted due to timeout.
	ErrExecutionTimeout ErrorType = 36

	// User was banned.
	ErrUserBanned ErrorType = 37

	// Unknown application.
	ErrUnknownApplication ErrorType = 38

	// Unknown user.
	ErrUnknownUser ErrorType = 39

	// Unknown group.
	ErrUnknownGroup ErrorType = 40

	// Additional signup required.
	ErrAdditionalSignupRequired ErrorType = 41

	// IP is not allowed.
	ErrIPNotAllowed ErrorType = 42

	// One of the parameters specified was missing or invalid
	//
	// Check the required parameters list and their format on a method
	// description page.
	ErrParam ErrorType = 100

	// Invalid application API ID
	//
	// Find the app in the administrated list in settings:
	// http://vk.com/apps?act=settings
	// And set the correct API_ID in the request.
	ErrParamAPIID   ErrorType = 101
	ErrLimits       ErrorType = 103 // Out of limits
	ErrNotFound     ErrorType = 104 // Not found
	ErrSaveFile     ErrorType = 105 // Couldn't save file
	ErrActionFailed ErrorType = 106 // Unable to process action

	// Invalid user id
	//
	// Make sure that you use a correct id. You can get an id using a screen
	// name with the utils.resolveScreenName method.
	ErrParamUserID  ErrorType = 113
	ErrParamAlbumID ErrorType = 114 // Invalid album id
	ErrParamServer  ErrorType = 118 // Invalid server
	ErrParamTitle   ErrorType = 119 // Invalid title
	ErrParamPhotos  ErrorType = 122 // Invalid photos
	ErrParamHash    ErrorType = 121 // Invalid hash
	ErrParamPhoto   ErrorType = 129 // Invalid photo
	ErrParamGroupID ErrorType = 125 // Invalid group id
	ErrParamPageID  ErrorType = 140 // Page not found
	ErrAccessPage   ErrorType = 141 // Access to page denied

	// The mobile number of the user is unknown.
	ErrMobileNotActivated ErrorType = 146

	// Application has insufficient funds.
	ErrInsufficientFunds ErrorType = 147

	// Access to the menu of the user denied.
	ErrAccessMenu ErrorType = 148

	// Invalid timestamp
	//
	// You may get a correct value with the utils.getServerTime method.
	ErrParamTimestamp ErrorType = 150
	ErrFriendsListID  ErrorType = 171 // Invalid list id

	// Reached the maximum number of lists.
	ErrFriendsListLimit ErrorType = 173

	// Cannot add user himself as friend.
	ErrFriendsAddYourself ErrorType = 174

	// Cannot add this user to friends as they have put you on their blacklist.
	ErrFriendsAddInEnemy ErrorType = 175

	// Cannot add this user to friends as you put him on blacklist.
	ErrFriendsAddEnemy ErrorType = 176

	// Cannot add this user to friends as user not found.
	ErrFriendsAddNotFound ErrorType = 177
	ErrParamNoteID        ErrorType = 180 // Note not found
	ErrAccessNote         ErrorType = 181 // Access to note denied
	ErrAccessNoteComment  ErrorType = 182 // You can't comment this note
	ErrAccessComment      ErrorType = 183 // Access to comment denied

	// Access to album denied
	//
	// Make sure you use correct ids (owner_id is always positive for users,
	// negative for communities) and the current user has access to the
	// requested content in the full version of the site.
	ErrAccessAlbum ErrorType = 200

	// Access to audio denied
	//
	// Make sure you use correct ids (owner_id is always positive for users,
	// negative for communities) and the current user has access to the
	// requested content in the full version of the site.
	ErrAccessAudio ErrorType = 201

	// Access to group denied
	//
	// Make sure that the current user is a member or admin of the community
	// (for closed and private groups and events).
	ErrAccessGroup ErrorType = 203

	// Access denied.
	ErrAccessVideo ErrorType = 204

	// Access denied.
	ErrAccessMarket ErrorType = 205

	// Access to wall's post denied.
	ErrWallAccessPost ErrorType = 210

	// Access to wall's comment denied.
	ErrWallAccessComment ErrorType = 211

	// Access to post comments denied.
	ErrWallAccessReplies ErrorType = 212

	// Access to status replies denied.
	ErrWallAccessAddReply ErrorType = 213

	// Access to adding post denied.
	ErrWallAddPost ErrorType = 214

	// Advertisement post was recently added.
	ErrWallAdsPublished ErrorType = 219

	// Too many recipients.
	ErrWallTooManyRecipients ErrorType = 220

	// User disabled track name broadcast.
	ErrStatusNoAudio ErrorType = 221

	// Hyperlinks are forbidden.
	ErrWallLinksForbidden ErrorType = 222

	// Too many replies.
	ErrWallReplyOwnerFlood ErrorType = 223

	// Too many ads posts.
	ErrWallAdsPostLimitReached ErrorType = 224

	// Donut is disabled.
	ErrDonutDisabled ErrorType = 225

	// Reaction can not be applied to the object.
	ErrLikesReactionCanNotBeApplied ErrorType = 232

	// Access to poll denied.
	ErrPollsAccess ErrorType = 250

	// Invalid answer id.
	ErrPollsAnswerID ErrorType = 252

	// Invalid poll id.
	ErrPollsPollID ErrorType = 251

	// Access denied, please vote first.
	ErrPollsAccessWithoutVote ErrorType = 253

	// Access to the groups list is denied due to the user's privacy settings.
	ErrAccessGroups ErrorType = 260

	// This album is full
	//
	// You need to delete the odd objects from the album or use another album.
	ErrAlbumFull   ErrorType = 300
	ErrAlbumsLimit ErrorType = 302 // Albums number limit is reached

	// Permission denied. You must enable votes processing in application
	// settings
	//
	// Check the app settings:
	//
	// 	http://vk.com/editapp?id={Your API_ID}&section=payments
	//
	ErrVotesPermission ErrorType = 500

	// Not enough votes.
	ErrVotes ErrorType = 503

	// Not enough money on owner's balance.
	ErrNotEnoughMoney ErrorType = 504

	// Permission denied. You have no access to operations specified with
	// given object(s).
	ErrAdsPermission ErrorType = 600

	// Permission denied. You have requested too many actions this day. Try
	// later.
	ErrWeightedFlood ErrorType = 601

	// Some part of the request has not been completed.
	ErrAdsPartialSuccess ErrorType = 602

	// Some ads error occurred.
	ErrAdsSpecific ErrorType = 603

	// Invalid domain.
	ErrAdsDomainInvalid ErrorType = 604

	// Domain is forbidden.
	ErrAdsDomainForbidden ErrorType = 605

	// Domain is reserved.
	ErrAdsDomainReserved ErrorType = 606

	// Domain is occupied.
	ErrAdsDomainOccupied ErrorType = 607

	// Domain is active.
	ErrAdsDomainActive ErrorType = 608

	// Domain app is invalid.
	ErrAdsDomainAppInvalid ErrorType = 609

	// Domain app is forbidden.
	ErrAdsDomainAppForbidden ErrorType = 610

	// Application must be verified.
	ErrAdsApplicationMustBeVerified ErrorType = 611

	// Application must be in domains list of site of ad unit.
	ErrAdsApplicationMustBeInDomainsList ErrorType = 612

	// Application is blocked.
	ErrAdsApplicationBlocked ErrorType = 613

	// Domain of type specified is forbidden in current office type.
	ErrAdsDomainTypeForbiddenInCurrentOffice ErrorType = 614

	// Domain group is invalid.
	ErrAdsDomainGroupInvalid ErrorType = 615

	// Domain group is forbidden.
	ErrAdsDomainGroupForbidden ErrorType = 616

	// Domain app is blocked.
	ErrAdsDomainAppBlocked ErrorType = 617

	// Domain group is not open.
	ErrAdsDomainGroupNotOpen ErrorType = 618

	// Domain group is not possible to be joined to adsweb.
	ErrAdsDomainGroupNotPossibleJoined ErrorType = 619

	// Domain group is blocked.
	ErrAdsDomainGroupBlocked ErrorType = 620

	// Domain group has restriction: links are forbidden.
	ErrAdsDomainGroupLinksForbidden ErrorType = 621

	// Domain group has restriction: excluded from search.
	ErrAdsDomainGroupExcludedFromSearch ErrorType = 622

	// Domain group has restriction: cover is forbidden.
	ErrAdsDomainGroupCoverForbidden ErrorType = 623

	// Domain group has wrong category.
	ErrAdsDomainGroupWrongCategory ErrorType = 624

	// Domain group has wrong name.
	ErrAdsDomainGroupWrongName ErrorType = 625

	// Domain group has low posts reach.
	ErrAdsDomainGroupLowPostsReach ErrorType = 626

	// Domain group has wrong class.
	ErrAdsDomainGroupWrongClass ErrorType = 627

	// Domain group is created recently.
	ErrAdsDomainGroupCreatedRecently ErrorType = 628

	// Object deleted.
	ErrAdsObjectDeleted ErrorType = 629

	// Lookalike request with same source already in progress.
	ErrAdsLookalikeRequestAlreadyInProgress ErrorType = 630

	// Max count of lookalike requests per day reached.
	ErrAdsLookalikeRequestsLimit ErrorType = 631

	// Given audience is too small.
	ErrAdsAudienceTooSmall ErrorType = 632

	// Given audience is too large.
	ErrAdsAudienceTooLarge ErrorType = 633

	// Lookalike request audience save already in progress.
	ErrAdsLookalikeAudienceSaveAlreadyInProgress ErrorType = 634

	// Max count of lookalike request audience saves per day reached.
	ErrAdsLookalikeSavesLimit ErrorType = 635

	// Max count of retargeting groups reached.
	ErrAdsRetargetingGroupsLimit ErrorType = 636

	// Domain group has active nemesis punishment.
	ErrAdsDomainGroupActiveNemesisPunishment ErrorType = 637

	// Cannot edit creator role.
	ErrGroupChangeCreator ErrorType = 700

	// User should be in club.
	ErrGroupNotInClub ErrorType = 701

	// Too many officers in club.
	ErrGroupTooManyOfficers ErrorType = 702

	// You need to enable 2FA for this action.
	ErrGroupNeed2fa ErrorType = 703

	// User needs to enable 2FA for this action.
	ErrGroupHostNeed2fa ErrorType = 704

	// Too many addresses in club.
	ErrGroupTooManyAddresses ErrorType = 706

	// "Application is not installed in community.
	ErrGroupAppIsNotInstalledInCommunity ErrorType = 711

	// Invite link is invalid - expired, deleted or not exists.
	ErrGroupInvalidInviteLink ErrorType = 714

	// This video is already added.
	ErrVideoAlreadyAdded ErrorType = 800

	// Comments for this video are closed.
	ErrVideoCommentsClosed ErrorType = 801

	// Can't send messages for users from blacklist.
	ErrMessagesUserBlocked ErrorType = 900

	// Can't send messages for users without permission.
	ErrMessagesDenySend ErrorType = 901

	// Can't send messages to this user due to their privacy settings.
	ErrMessagesPrivacy ErrorType = 902

	// Value of ts or pts is too old.
	ErrMessagesTooOldPts ErrorType = 907

	// Value of ts or pts is too new.
	ErrMessagesTooNewPts ErrorType = 908

	// Can't edit this message, because it's too old.
	ErrMessagesEditExpired ErrorType = 909

	// Can't sent this message, because it's too big.
	ErrMessagesTooBig ErrorType = 910

	// Keyboard format is invalid.
	ErrMessagesKeyboardInvalid ErrorType = 911

	// This is a chat bot feature, change this status in settings.
	ErrMessagesChatBotFeature ErrorType = 912

	// Too many forwarded messages.
	ErrMessagesTooLongForwards ErrorType = 913

	// Message is too long.
	ErrMessagesTooLongMessage ErrorType = 914

	// You don't have access to this chat.
	ErrMessagesChatUserNoAccess ErrorType = 917

	// You can't see invite link for this chat.
	ErrMessagesCantSeeInviteLink ErrorType = 919

	// Can't edit this kind of message.
	ErrMessagesEditKindDisallowed ErrorType = 920

	// Can't forward these messages.
	ErrMessagesCantFwd ErrorType = 921

	// Can't delete this message for everybody.
	ErrMessagesCantDeleteForAll ErrorType = 924

	// You are not admin of this chat.
	ErrMessagesChatNotAdmin ErrorType = 925

	// Chat does not exist.
	ErrMessagesChatNotExist ErrorType = 927

	// You can't change invite link for this chat.
	ErrMessagesCantChangeInviteLink ErrorType = 931

	// Your community can't interact with this peer.
	ErrMessagesGroupPeerAccess ErrorType = 932

	// User not found in chat.
	ErrMessagesChatUserNotInChat ErrorType = 935

	// Contact not found.
	ErrMessagesContactNotFound ErrorType = 936

	// Message request already send.
	ErrMessagesMessageRequestAlreadySend ErrorType = 939

	// Too many posts in messages.
	ErrMessagesTooManyPosts ErrorType = 940

	// Cannot pin one-time story.
	ErrMessagesCantPinOneTimeStory ErrorType = 942

	// Cannot use this intent.
	ErrMessagesCantUseIntent ErrorType = 943

	// Limits overflow for this intent.
	ErrMessagesLimitIntent ErrorType = 944

	// Chat was disabled.
	ErrMessagesChatDisabled ErrorType = 945

	// Chat not support.
	ErrMessagesChatNotSupported ErrorType = 946

	// Can't add user to chat, because user has no access to group.
	ErrMessagesMemberAccessToGroupDenied ErrorType = 947

	// Can't edit pinned message yet.
	ErrMessagesEditPinned ErrorType = 949

	// Can't send message, reply timed out.
	ErrMessagesReplyTimedOut ErrorType = 950

	// You can't access donut chat without subscription.
	ErrMessagesAccessDonutChat ErrorType = 962

	// This user can't be added to the work chat, as they aren't an employe.
	ErrMessagesAccessWorkChat ErrorType = 967

	// Message cannot be forwarded.
	ErrMessagesCantForwarded ErrorType = 969

	// Cannot pin an expiring message.
	ErrMessagesPinExpiringMessage ErrorType = 970

	// Invalid phone number.
	ErrParamPhone ErrorType = 1000

	// This phone number is used by another user.
	ErrPhoneAlreadyUsed ErrorType = 1004

	// Too many auth attempts, try again later.
	ErrAuthFloodError ErrorType = 1105

	// Processing.. Try later.
	ErrAuthDelay ErrorType = 1112

	// Anonymous token has expired.
	ErrAnonymousTokenExpired ErrorType = 1114

	// Anonymous token is invalid.
	ErrAnonymousTokenInvalid ErrorType = 1116

	// Access token has expired.
	ErrAuthAccessTokenHasExpired ErrorType = 1117

	// Anonymous token ip mismatch.
	ErrAuthAnonymousTokenIPMismatch ErrorType = 1118

	// Invalid document id.
	ErrParamDocID ErrorType = 1150

	// Access to document deleting is denied.
	ErrParamDocDeleteAccess ErrorType = 1151

	// Invalid document title.
	ErrParamDocTitle ErrorType = 1152

	// Access to document is denied.
	ErrParamDocAccess ErrorType = 1153

	// Original photo was changed.
	ErrPhotoChanged ErrorType = 1160

	// Too many feed lists.
	ErrTooManyLists ErrorType = 1170

	// This achievement is already unlocked.
	ErrAppsAlreadyUnlocked ErrorType = 1251

	// Subscription not found.
	ErrAppsSubscriptionNotFound ErrorType = 1256

	// Subscription is in invalid status.
	ErrAppsSubscriptionInvalidStatus ErrorType = 1257

	// Invalid screen name.
	ErrInvalidAddress ErrorType = 1260

	// Catalog is not available for this user.
	ErrCommunitiesCatalogDisabled ErrorType = 1310

	// Catalog categories are not available for this user.
	ErrCommunitiesCategoriesDisabled ErrorType = 1311

	// Too late for restore.
	ErrMarketRestoreTooLate ErrorType = 1400

	// Comments for this market are closed.
	ErrMarketCommentsClosed ErrorType = 1401

	// Album not found.
	ErrMarketAlbumNotFound ErrorType = 1402

	// Item not found.
	ErrMarketItemNotFound ErrorType = 1403

	// Item already added to album.
	ErrMarketItemAlreadyAdded ErrorType = 1404

	// Too many items.
	ErrMarketTooManyItems ErrorType = 1405

	// Too many items in album.
	ErrMarketTooManyItemsInAlbum ErrorType = 1406

	// Too many albums.
	ErrMarketTooManyAlbums ErrorType = 1407

	// Item has bad links in description.
	ErrMarketItemHasBadLinks ErrorType = 1408

	// Extended market not enabled.
	ErrMarketShopNotEnabled ErrorType = 1409

	// Grouping items with different properties.
	ErrMarketGroupingItemsWithDifferentProperties ErrorType = 1412

	// Grouping already has such variant.
	ErrMarketGroupingAlreadyHasSuchVariant ErrorType = 1413

	// Variant not found.
	ErrMarketVariantNotFound ErrorType = 1416

	// Property not found.
	ErrMarketPropertyNotFound ErrorType = 1417

	// Grouping must have two or more items.
	ErrMarketGroupingMustContainMoreThanOneItem ErrorType = 1425

	// Item must have distinct properties.
	ErrMarketGroupingItemsMustHaveDistinctProperties ErrorType = 1426

	// Cart is empty.
	ErrMarketOrdersNoCartItems ErrorType = 1427

	// Specify width, length, height and weight all together.
	ErrMarketInvalidDimensions ErrorType = 1429

	// VK Pay status can not be changed.
	ErrMarketCantChangeVkpayStatus ErrorType = 1430

	// Market was already enabled in this group.
	ErrMarketShopAlreadyEnabled ErrorType = 1431

	// Market was already disabled in this group.
	ErrMarketShopAlreadyDisabled ErrorType = 1432

	// Invalid image crop format.
	ErrMarketPhotosCropInvalidFormat ErrorType = 1433

	// Crop bottom right corner is outside of the image.
	ErrMarketPhotosCropOverflow ErrorType = 1434

	// Crop size is less than the minimum.
	ErrMarketPhotosCropSizeTooLow ErrorType = 1435

	// Market not enabled.
	ErrMarketNotEnabled ErrorType = 1438

	// Cart is empty.
	ErrMarketCartEmpty ErrorType = 1427

	// Specify width, length, height and weight all together.
	ErrMarketSpecifyDimensions ErrorType = 1429

	// VK Pay status can not be changed.
	ErrVKPayStatus ErrorType = 1430

	// Market was already enabled in this group.
	ErrMarketAlreadyEnabled ErrorType = 1431

	// Market was already disabled in this group.
	ErrMarketAlreadyDisabled ErrorType = 1432

	// Main album can not be hidden.
	ErrMainAlbumCantHidden ErrorType = 1446

	// Story has already expired.
	ErrStoryExpired ErrorType = 1600

	// Incorrect reply privacy.
	ErrStoryIncorrectReplyPrivacy ErrorType = 1602

	// Card not found.
	ErrPrettyCardsCardNotFound ErrorType = 1900

	// Too many cards.
	ErrPrettyCardsTooManyCards ErrorType = 1901

	// Card is connected to post.
	ErrPrettyCardsCardIsConnectedToPost ErrorType = 1902

	// Servers number limit is reached.
	ErrCallbackServersLimit ErrorType = 2000

	// Stickers are not purchased.
	ErrStickersNotPurchased ErrorType = 2100

	// Too many favorite stickers.
	ErrStickersTooManyFavorites ErrorType = 2101

	// Stickers are not favorite.
	ErrStickersNotFavorite ErrorType = 2102

	// Specified link is incorrect (can't find source).
	ErrWallCheckLinkCantDetermineSource ErrorType = 3102

	// Recaptcha needed.
	ErrRecaptcha ErrorType = 3300

	// Phone validation needed.
	ErrPhoneValidation ErrorType = 3301

	// Password validation needed.
	ErrPasswordValidation ErrorType = 3302

	// Otp app validation needed.
	ErrOtpAppValidation ErrorType = 3303

	// Email confirmation needed.
	ErrEmailConfirmation ErrorType = 3304

	// Assert votes.
	ErrAssertVotes ErrorType = 3305

	// Token extension required.
	ErrTokenExtension ErrorType = 3609

	// User is deactivated.
	ErrUserDeactivated ErrorType = 3610

	// Service is deactivated for user.
	ErrServiceDeactivated ErrorType = 3611

	// Can't set AliExpress tag to this type of object.
	ErrAliExpressTag ErrorType = 3800

	// Invalid upload response.
	ErrInvalidUploadResponse ErrorType = 5701

	// Invalid upload hash.
	ErrInvalidUploadHash ErrorType = 5702

	// Invalid upload user.
	ErrInvalidUploadUser ErrorType = 5703

	// Invalid upload group.
	ErrInvalidUploadGroup ErrorType = 5704

	// Invalid crop data.
	ErrInvalidCropData ErrorType = 5705

	// To small avatar.
	ErrToSmallAvatar ErrorType = 5706

	// Photo not found.
	ErrPhotoNotFound ErrorType = 5708

	// Invalid Photo.
	ErrInvalidPhoto ErrorType = 5709

	// Invalid hash.
	ErrInvalidHash ErrorType = 5710
)

// ErrorSubtype is the subtype of an error.
type ErrorSubtype int

// Error returns the message of a ErrorSubtype.
func (e ErrorSubtype) Error() string {
	return fmt.Sprintf("api: error with subcode %d", e)
}

// Error struct VK.
type Error struct {
	Code       ErrorType    `json:"error_code"`
	Subcode    ErrorSubtype `json:"error_subcode"`
	Message    string       `json:"error_msg"`
	Text       string       `json:"error_text"`
	CaptchaSID string       `json:"captcha_sid"`
	CaptchaImg string       `json:"captcha_img"`

	// In some cases VK requires to request action confirmation from the user
	// (for Standalone apps only). Following error will be returned:
	//
	// Error code: 24
	// Error text: Confirmation required
	//
	// Following parameter is transmitted with the error message as well:
	//
	// confirmation_text – text of the message to be shown in the default
	// confirmation window.
	//
	// The app should display the default confirmation window with text from
	// confirmation_text and two buttons: "Continue" and "Cancel". If user
	// confirms the action repeat the request with an extra parameter:
	// confirm = 1.
	//
	// See https://vk.com/dev/need_confirmation
	ConfirmationText string `json:"confirmation_text"`

	// In some cases VK requires a user validation procedure. . As a result
	// starting from API version 5.0 (for the older versions captcha_error
	// will be requested) following error will be returned as a reply to any
	// API request:
	//
	// Error code: 17
	// Error text: Validation Required
	//
	// Following parameter is transmitted with an error message:
	// redirect_uri – a special address to open in a browser to pass the
	// validation procedure.
	//
	// After passing the validation a user will be redirected to the service
	// page:
	//
	// https://oauth.vk.com/blank.html#{Data required for validation}
	//
	// In case of successful validation following parameters will be
	// transmitted after #:
	//
	// https://oauth.vk.com/blank.html#success=1&access_token={NEW USER TOKEN}&user_id={USER ID}
	//
	// If a token was not received by https a new secret will be transmitted
	// as well.
	//
	// In case of unsuccessful validation following address is transmitted:
	//
	// https://oauth.vk.com/blank.html#fail=1
	//
	// See https://vk.com/dev/need_validation
	RedirectURI   string                    `json:"redirect_uri"`
	RequestParams []object.BaseRequestParam `json:"request_params"`
}

// Error returns the message of a Error.
func (e Error) Error() string {
	return "api: " + e.Message
}

// Is unwraps its first argument sequentially looking for an error that matches
// the second.
func (e Error) Is(target error) bool {
	var tError *Error
	if errors.As(target, &tError) {
		return e.Code == tError.Code && e.Message == tError.Message
	}

	var tErrorType ErrorType
	if errors.As(target, &tErrorType) {
		return e.Code == tErrorType
	}

	return false
}

// ExecuteError struct.
//
// TODO: v3 Code is ErrorType.
type ExecuteError struct {
	Method string `json:"method"`
	Code   int    `json:"error_code"`
	Msg    string `json:"error_msg"`
}

// ExecuteErrors type.
type ExecuteErrors []ExecuteError

// Error returns the message of a ExecuteErrors.
func (e ExecuteErrors) Error() string {
	return fmt.Sprintf("api: execute errors (%d)", len(e))
}

// InvalidContentType type.
type InvalidContentType struct {
	ContentType string
}

// Error returns the message of a InvalidContentType.
func (e InvalidContentType) Error() string {
	return "api: invalid content-type"
}

// UploadError type.
type UploadError struct {
	Err      string `json:"error"`
	Code     int    `json:"error_code"`
	Descr    string `json:"error_descr"`
	IsLogged bool   `json:"error_is_logged"`
}

// Error returns the message of a UploadError.
func (e UploadError) Error() string {
	if e.Err != "" {
		return "api: " + e.Err
	}

	return fmt.Sprintf("api: upload code %d", e.Code)
}

// AdsError struct.
type AdsError struct {
	Code ErrorType `json:"error_code"`
	Desc string    `json:"error_desc"`
}

// Error returns the message of a AdsError.
func (e AdsError) Error() string {
	return "api: " + e.Desc
}

// Is unwraps its first argument sequentially looking for an error that matches
// the second.
func (e AdsError) Is(target error) bool {
	var tAdsError *AdsError
	if errors.As(target, &tAdsError) {
		return e.Code == tAdsError.Code && e.Desc == tAdsError.Desc
	}

	var tErrorType ErrorType
	if errors.As(target, &tErrorType) {
		return e.Code == tErrorType
	}

	return false
}

// AuthSilentTokenError struct.
type AuthSilentTokenError struct {
	Token       string    `json:"token"`
	Code        ErrorType `json:"code"`
	Description string    `json:"description"`
}

// Error returns the description of a AuthSilentTokenError.
func (e AuthSilentTokenError) Error() string {
	return "api: " + e.Description
}

// Is unwraps its first argument sequentially looking for an error that matches
// the second.
func (e AuthSilentTokenError) Is(target error) bool {
	var tError *AuthSilentTokenError
	if errors.As(target, &tError) {
		return e.Code == tError.Code && e.Description == tError.Description
	}

	var tErrorType ErrorType
	if errors.As(target, &tErrorType) {
		return e.Code == tErrorType
	}

	return false
}
