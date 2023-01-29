# v1.26.0

## New features

- irc: Allow substitution of bot's nick in RunCommands (irc) (#1890)
- matrix: Add Matrix username spoofing (#1875)

## Enhancements

- general: Update dependencies (#1951)
- mattermost: Remove mattermost 5 support (#1936)
- mumble: Implement sending of EventJoinLeave both to and from Mumble (#1915)
- whatsappmulti: Improve attachment handling (whatsapp) (#1928)
- whatsappmulti: Handle incoming document captions from whatsapp (#1935)

## Bugfix

- irc: Fix empty messages on IRC (#1897)
- telegram: Fix message html entities escaping when sending to Telegram (#1855)
- telegram/slack: Fix error messages in telegram and slack bridges (#1862)
- telegram: Fix telegram attachment comment formatting and escaping (#1920)
- telegram: Make the cgo lottie a build tag (-tag cgolottie) (#1955)
- whatsappmulti: Update dependencies and fix whatsmeow API changes (#1887)
- whatsappmulti: Fix the "Someone" nickname problem (whatsapp) (#1931)

This release couldn't exist without the following contributors:
@s3lph, @sas1024, @Glandos, @jx11r, @Lucki, @BuckarooBanzay, @ilmaisin, @Kufat

# v1.25.2

## Enhancements

- general: Update dependencies (#1851,#1841)
- mattermost: Support mattermost v7.x (#1852)

## Bugfix

- discord: Fix Unwanted join notifications from one Discord server to another (#1612)
- discord: Ignore events from other guilds, add nosendjoinpart support (#1846)

This release couldn't exist without the following contributors:
@wlcx

# v1.25.1

## Enhancements

- matrix: Add KeepQuotedReply option for matrix to fix regression (#1823)
- slack: Improve Slack attachments formatting (slack) (#1807)

## Bugfix

- general: Update dependencies (#1813,#1822,#1833)
- mattermost: Add space between filename and URL (mattermost). Fixes #1820
- matrix: Update matterbridge/gomatrix. Fixes #1772 (#1803)
- telegram: Do not modify .webm files (telegram). Fixes #17**88 (#1802)
- telegram: Do not apply any markup to URL entities (telegram) (#1808)
- telegram: Fix telegram message deletion request (#1818)
- vk: Fix UploadMessagesPhoto for vk community chat (vk) (#1812)

This release couldn't exist without the following contributors:
@bd808, @chugunov, @sas1024, @SevereCloud, @ValdikSS

# v1.25.0

## Breaking changes

- whatsapp: deprecated, the library <https://github.com/Rhymen/go-whatsapp> isn't maintained anymore.
We're switching to <https://github.com/tulir/whatsmeow> but as this uses a GPL3 licensed library we can't provide you with binaries.
You'll need to build it yourself. More information about this can be found here: <https://github.com/42wim/matterbridge#building-with-whatsapp-beta-multidevice-support>

## New features

- whatsappmulti: whatsapp multidevice support added - more info <https://github.com/42wim/matterbridge#building-with-whatsapp-beta-multidevice-support>
- general: Add Dockerfile_whatsappmulti for building with WhatsApp Multi-Device support (Whatsmeow) (#1774)
- telegram: Add UseFullName option (telegram) (#1777)
- slack: Use slack real name as user name (slack) (#1775)

## Enhancements

- general: Ignore sending file with comment, if comment contains IgnoreMessages value (#1783)
- general: Update dependencies (#1784)
- irc: Update lrstanley/girc dep (#1773)
- slack: Preserve threading for messages with files (slack) (#1781)
- telegram: Preserve threading from telegram replies (telegram) (#1776)
- telegram: Multiple media in one message (telegram) (#1779)
- whatsapp: Add whatsapp deprecation warning (#1792)

## Bugfix

- discord: Change discord non-native threading behaviour (discord) (#1791)

This release couldn't exist without the following contributors:
@sas1024, @tpxtron

# v1.24.1

## Enhancements

- discord: Switch to discordgo upstream again (#1759)
- general: Update dependencies and vendor (#1761)
- general: Create inmessage-logger.tengo (#1688) (#1747)
- general: Add OpenRC service file (#1746)
- irc: Refactor utf-8 conversion (irc) (#1767)

## Bugfixes

- irc: Fix panic in irc. Closes #1751 (#1760)
- mumble: Implement a workaround to signal Opus support (mumble) (#1764)
- telegram: Fix for complex-formatted Telegram text (#1765)
- telegram: Fix Telegram channel title in forwards (#1753)
- telegram: Fix Telegram Problem (unforwarded formatting and skipping of linebreaks) (#1749)

This release couldn't exist without the following contributors:
@s3lph, @ValdikSS, @reckel-jm, @CyberTailor

# v1.24.0

## New features

- harmony: new protocol added: Add support for Harmony (#1656)
- irc: Allow binding to IP on IRC (#1640)
- irc: Add support for client certificate (irc) (#1710)
- mattermost: Add UseUsername option (mattermost). Fixes #1665 (#1714)
- mattermost: Add support for using ID in channel config (mattermost) (#1715)
- matrix: Reply support for Matrix (#1664)
- telegram: Add Telegram Bot Command /chatId (telegram) (#1703)

## Enhancements

- general: Update dependencies/vendor (#1659)
- discord: Add more debug options for discord (#1712)
- docker: Use Alpine stable again in Dockerfile (#1643)
- mattermost: Log eventtype in debug (mattermost) (#1676)
- mattermost: Add more ignore debug messages (mattermost) (#1678)
- slack: Add support for deleting files from slack to discord. Fixes #1705 (#1709)
- telegram: Add support for code blocks in telegram (#1650)
- telegram: Update telegram-bot-api to v5 (#1660)
- telegram: Add comments to messages (telegram) (#1652)
- telegram: Add support for sender_chat (telegram) (#1677)
- vk: Remove GroupID (vk) (#1668)

## Bugfix

- mattermost: Use current parentID if rootId is not set (mattermost) (#1675)
- matrix: Make HTMLDisable work correct (matrix) (#1716)
- whatsapp: Make EditSuffix option actually work (whatsapp). Fixes #1510 (#1728)

This release couldn't exist without the following contributors:
@DavyJohnesev, @GoliathLabs, @pontaoski, @PeGaSuS-Coder, @dependabot[bot], @vpzomtrrfrt, @SevereCloud, @soloam, @YashRE42, @danwalmsley, @SuperSandro2000, @inzanity

# v1.23.2

If you're running whatsapp you should update.

## Bugfix

- whatsapp: Update go-whatsapp version (#1630)

This release couldn't exist without the following contributors:
@snikpic

# v1.23.1

If you're running mattermost 6 you should update.

## Bugfix

- mattermost: Do not check cache on deleted messages (mattermost). Fixes #1555 (#1624)
- mattermost: Fix crash on users updating info. Update matterclient dep. Fixes #1617
- matrix: Keep the logger on a disabled bridge. Fixes #1616 (#1621)
- msteams: Fix panic in msteams. Fixes #1588 (#1622)
- xmpp: Do not fail on no avatar data (xmpp) #1529 (#1627)
- xmpp: Use a new msgID when replacing messages (xmpp). Fixes #1584 (#1623)
- zulip: Add better error handling on Zulip (#1589)

This release couldn't exist without the following contributors:
@Polynomdivision, @minecraftchest1, @alexmv

# v1.23.0

## New features

- irc: Add UserName and RealName options for IRC (#1590)
- mattermost: Add support for mattermost v6
- nctalk: Add support for separate display name (nctalk) (#1506)
- xmpp: Add support for anonymous connection (xmpp) (#1548)

## Enhancements

- general: Update vendored libraries
- docker: Use github actions to build dockerhub/ghcr.io images
- docker: Update GH actions to multi arch (arm64) (#1614)
- telegram: Convert .tgs with go libraries (and cgo) (telegram) (#1569)

## Bugfix

- mumble: Remove newline character in bridge multiline messages (mumble) (#1572)
- slack: Add space before file upload comment (slack) (#1554)
- slack: Invalidate user in cache on user change event (#1604)
- xmpp: Fix XMPP parseNick function (#1547)

This release couldn't exist without the following contributors:
@powerjungle, @gary-kim, @KingPin, @Benau, @keenan-v1, @tytan652, @KidA001,@minecraftchest1, @irydacea

# v1.22.3

## Bugfixes

- whatsapp: Update Rhymen/go-whatsapp module to latest master (2b8a3e9b8aa2) (#1518)

This release couldn't exist without the following contributors:
@nathanaelhoun

# v1.22.2

## Enhancements

- general: Add a MessageClipped option to set your own clipped message. Closes #1359 (#1487)
- discord: Add AllowMention to restrict allowed mentions (#1462)
- matrix: Add MxId/Token login option for Matrix (#1438)
- nctalk: Support sending file URLs (nctalk) (#1489)
- nctalk: Add support for message deletion (nctalk) (#1492)
- whatsapp: Handle document messages (whatsapp) (#1475)

## Bugfixes

- general: Update vendored libs
- matrix: Fix content body issue for redactions (matrix) (#1496)
- telegram: Add libwebp-dev to tgs.Dockerfile fixes Telegram sticker to WebP rendering (#1476)
- whatsapp: Rename .jpe files to .jpg Fixes #1463 (whatsapp) (#1485)
- whatsapp: Fix crash on encountering VideoMessage (whatsapp) (#1483)

This release couldn't exist without the following contributors:
@AvinashReddy3108, @chrisbobbe, @jaywink, @Funatiker, @computeronix, @alexandregv, @gary-kim, @SuperSandro2000

# v1.22.1

## Enhancements

- rocketchat: Handle Rocket.Chat attachments (#1395)
- telegram: Adding caption to send telegram images. Fixes #1357 (#1358)
- whatsapp: Set ogg as default audiomessage when none found (whatsapp). Fixes #1427 (#1431)

## Bugfixes

- discord: Declare GUILD_MEMBERS privileged intent (discord) (#1428)
- telegram: Check rune length instead of bytes (telegram). Fixes #1409 (#1412)
- telegram: Make lottie_convert work on platforms without /dev/stdout (#1424)
- xmpp: Fix panic when the webhook fails (xmpp) (#1401)
- xmpp: Fix webhooks for channels with special characters (xmpp) (#1405)

This release couldn't exist without the following contributors:
@BenWiederhake, @powerjungle, @qaisjp, @Humorhenker, @Polynomdivision, @tadeokondrak, @PeGaSuS-Coder, @Millesimus, @jlu5

# v1.22.0

Discord users using autowebhooks are encouraged to upgrade to this release.

## New features

- vk: new protocol added: Add vk support (#1245)
- xmpp: Allow the XMPP bridge to use slack compatible webhooks (xmpp) (#1364)

## Enhancements

- telegram: Rename .oga audio files to .ogg (telegram) (#1349)
- telegram: Add jpe as valid image filename extension (telegram) (#1360)
- discord: Add an even more debug option (discord) (#1368)
- general: Update vendor (#1384)

## Bugfixes

- discord: Pick up all the webhooks (discord) (#1383). Fixes #1353

This release couldn't exist without the following contributors:
@ivanik7, @Polynomdivision, @PeterDaveHello, @Humorhenker, @qaisjp

# v1.21.0

## Breaking Changes

- discord: Remove WebhookURL support (discord) (#1323)

`WebhookURL` global setting for discord is removed and will quit matterbridge.
New `AutoWebhooks=true` setting, which will automatically use (and create, if they do not exist) webhooks inside specific channels. This only works if the bot has Manage Webhooks permission in bridged channels (global permission or as a channel permission override). Backwards compatibility with channel-specific webhooks. More info [here](https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample#L862).

## New features

- discord: Create webhooks automatically (#1323)
- discord: Add threading support with token (discord) (#1342)
- irc: Join on invite (irc). Fixes #1231 (#1306)
- irc: Add support for stateless bridging via draft/relaymsg (irc) (#1339)
- whatsapp: Add support for deleting messages (whatsapp) (#1316)
- whatsapp: Handle video downloads (whatsapp) (#1316)
- whatsapp: Handle audio downloads (whatsapp) (#1316)

## Enhancements

- general: Parse fencedcode in ParseMarkdown. Fixes #1127 (#1329)
- discord: Refactor guild finding code (discord) (#1319)
- discord: Add a prefix handler for unthreaded messages (discord) (#1346)
- irc: Add support for irc to irc notice (irc). Fixes #754 (#1305)
- irc: Make handlers run async (irc) (#1325)
- matrix: Show mxids in case of clashing usernames (matrix) (#1309)
- matrix: Implement ratelimiting (matrix). Fixes #1238 (#1326)
- matrix: Mark messages as read (matrix). Fixes #1317 (#1328)
- nctalk: Update go-nc-talk (nctalk) (#1333)
- rocketchat: Update rocketchat vendor (#1327)
- tengo: Add UserID to RemoteNickFormat and Tengo (#1308)
- whatsapp: Retry until we have contacts (whatsapp). Fixes #1122 (#1304)
- whatsapp: Refactor/cleanup code (whatsapp)
- whatsapp: Refactor handleTextMessage (whatsapp)
- whatsapp: Refactor image downloads (whatsapp)
- whatsapp: Rename jfif to jpg (whatsapp). Fixes #1292

## Bugfix

- discord: Reject cross-channel message references (discord) (#1345)
- mumble: Add nil checks to text message handling (mumble) (#1321)

This release couldn't exist without the following contributors:
@nightmared, @qaisjp, @jlu5, @wschwab, @gary-kim, @s3lph, @JeremyRand

# v1.20.0

## Breaking

- matrix: Send the display name instead of the user name (matrix) (#1282)  
  Matrix now sends the displayname if set instead of the username. If you want to keep the username, add `UseUsername=true` to your matrix config. <https://github.com/42wim/matterbridge/wiki/Settings#useusername-1>
- discord: Disable webhook editing (discord) (#1296)  
  Because of issues with ratelimiting of webhook editing, this feature is now disabled. If you have multiple discord channels you bridge, you'll need to add a `webhookURL` to the `[gateway.inout.options]`. See <https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample#L1864-L1870> for an example.

## New features

- general: Allow tengo to drop messages using msgDrop (#1272)
- general: Update libraries (whatsapp,markdown,mattermost,ssh-chat)
- irc: Add PingDelay option (irc) (#1269)
- matrix: Allow message edits on matrix (#1286)
- xmpp: add NoTLS option to allow plaintext XMPP connections (#1288)

## Enhancements

- discord: Edit messages via webhook (1287)
- general: Add extra debug to log time spent sending a message per bridge (#1299)

This release couldn't exist without the following contributors:
@nightmared, @zhoreeq

# v1.19.0

## New features

- mumble: new protocol added: Add Mumble support (#1245)
- nctalk: Add support for downloading files (nctalk) (#1249)
- nctalk: Append a suffix if user is a guest user (nctalk) (#1250)

## Enhancements

- irc: Add even more debug for irc (#1266)
- matrix: Add username formatting for all events (matrix) (#1233)
- matrix: Permit uploading files of other mimetypes (#1237)
- whatsapp: Use vendored whatsapp version (#1258)
- whatsapp: Add username for images from WhatsApp (#1232)

This release couldn't exist without the following contributors:
@Dellle, @42wim, @gary-kim, @s3lph, @BenWiederhake

# v1.18.3

## Enhancements

- nctalk: Add TLSConfig to nctalk (#1195)
- whatsapp: Handle broadcasts as groups in Whatsapp #1213
- matrix: switch to upstream gomatrix #1219
- api: support multiple websocket clients #1205

## Bugfix

- general: update vendor
- zulip: Check location of avatarURL (zulip). Fixes #1214 (#1227)
- nctalk: Fix issue with too many open files #1223
- nctalk: Fix mentions #1222
- nctalk: Fix message replays #1220

This release couldn't exist without the following contributors:
@gary-kim, @tilosp, @NikkyAI, @escoand, @42wim

# v1.18.2

## Bugfix

- zulip: Fix error loop (zulip) (#1210)
- whatsapp: Update whatsapp vendor and fix a panic (#1209)

This release couldn't exist without the following contributors:
@SuperSandro2000, @42wim

# v1.18.1

## New features

- telegram: Support Telegram animated stickers (tgs) format (#1173). See https://github.com/42wim/matterbridge/wiki/Settings#mediaConverttgs for more info

## Enhancements

- matrix: Remove HTML formatting for push messages (#1188) (#1189)
- mattermost: Use mattermost v5 module (#1192)

## Bugfix

- whatsapp: Handle panic in whatsapp. Fixes #1180 (#1184)
- nctalk: Fix Nextcloud Talk connection failure (#1179)
- matrix: Sleep when ratelimited on joins (matrix). Fixes #1201 (#1206)

This release couldn't exist without the following contributors:
@42wim, @BenWiederhake, @Dellle, @gary-kim

# v1.18.0

## New features

- nctalk: new protocol added. Add Nextcloud Talk support #1167
- general: Add an option to log into a file rather than stdout (#1168)
- api: Add websocket to API (#970)

## Enhancements

- telegram: Fix MarkdownV2 support in Telegram (#1169)
- whatsapp: Reload user information when a new contact is detected (whatsapp) (#1160)
- api: Add sane RemoteNickFormat default for API (#1157)
- irc: Skip gIRC built-in rate limiting (irc) (#1164)
- irc: Only colour IRC nicks if there is one. (#1161)
- docker: Combine runs to one layer (#1151)

## Bugfix

- general: Update dependencies for 1.18.0 release (#1175)

Discord users are encouraged to upgrade, this release works with the move to the discord.com domain.

This release couldn't exist without the following contributors:
@42wim, @jlu5, @qaisjp, @TheHolyRoger, @SuperSandro2000, @gary-kim, @z3bra, @greenx, @haykam821, @nathanaelhoun

# v1.17.5

## Enhancements

- irc: Add StripMarkdown option (irc). (#1145)
- general: Increase debug logging with function,file and linenumber (#1147)
- general: Update Dockerfile so inotify works (#1148)
- matrix: Add an option to disable sending HTML to matrix. Fixes #1022 (#1135)
- xmpp: Implement xep-0245 (xmpp). Closes #1137 (#1144)

## Bugfix

- discord: Fix #1120: replaceAction "\_" crash (discord) (#1121)
- discord: Fix #1049: missing space before embeds (discord) (#1124)
- discord: Fix webhook EventUserAction messages being skipped (discord) (#1133)
- matrix: Avoid creating invalid url when the user doesn't have an avatar (matrix) (#1130)
- msteams: Ignore non-user messages (msteams). Fixes #1141 (#1149)
- slack: Do not use webhooks when token is configured (slack) (fixes #1123) (#1134)
- telegram: Fix forward from hidden users (telegram). Closes #1131 (#1143)
- xmpp: Prevent re-requesting avatar data (xmpp) (#1117)

This release couldn't exist without the following contributors:
@qaisjp, @xnaas, @42wim, @Polynomdivision, @tfve

# v1.17.4

## Bugfix

- general: Lowercase account names. Fixes #1108 (#1110)
- msteams: Remove panics and retry polling on failure (msteams). Fixes #1104 (#1105
- whatsapp: Update Rhymen/go-whatsapp. Fixes #1107 (#1109) (make whatsapp working again)
- discord: Add an ID cache (discord). Fixes #1106 (#1111) (fix delete/edits with webhooks)

# v1.17.3

## Enhancements

- xmpp: Implement User Avatar spoofing of XMPP users #1090
- rocketchat: Relay Joins/Topic changes in RocketChat bridge (#1085)
- irc: Add JoinDelay option (irc). Fixes #1084 (#1098)
- slack: Clip too long messages on 3000 length (slack). Fixes #1081 (#1102)

## Bugfix

- general: Fix the behavior of ShowTopicChange and SyncTopic (#1086)
- slack: Prevent image/message looping (slack). Fixes #1088 (#1096)
- whatsapp: Ignore non-critical errors (whatsapp). Fixes #1094 (#1100)
- irc: Add extra space before colon in attachments (irc). Fixes #1089 (#1101)

This release couldn't exist without the following contributors:
@42wim, @ldruschk, @qaisjp, @Polynomdivision

# v1.17.2

## Enhancements

- slack: Update vendor slack-go/slack (#1068)
- general: Update vendor d5/tengo (#1066)
- general: Clarify terminology used in mapping group chat IDs to channels in config (#1079)

## Bugfix

- whatsapp: Update Rhymen/go-whatsapp vendor and whatsapp version (#1078). Fixes Media upload #1074
- whatsapp: Reset start timestamp on reconnect (whatsapp). Fixes #1059 (#1064)

This release couldn't exist without the following contributors:
@42wim, @jheiselman

# v1.17.1

## Enhancements

- docker: Remove build dependencies from final image (multistage build) #1057

## Bugfix

- general: Don't transmit typing events from ourselves #1056
- general: Add support for build tags #1054
- discord: Strip extra info from emotes (discord) #1052
- msteams: fix macos build: Update vendor yaegashi/msgraph.go to v0.1.2 #1036
- whatsapp: Update client version whatsapp. Fixes #1061 #1062

This release couldn't exist without the following contributors:
@awigen, @qaisjp, @42wim

# v1.17.0

## New features

- msteams: new protocol added. Add initial Microsoft Teams support #967
  See https://github.com/42wim/matterbridge/wiki/MS-Teams-setup for a complete walkthrough
- discord: Add ability to procure avatars from the destination bridge #1000
- matrix: Add support for avatars from matrix. #1007
- general: support JSON and YAML config formats #1045

## Enhancements

- discord: Check only bridged channels for PermManageWebhooks #1001
- irc: Be less lossy when throttling IRC messages #1004
- keybase: updated library #1002, #1019
- matrix: Rebase gomatrix vendor with upstream #1006
- slack: Use upstream slack-go/slack again #1018
- slack: Ignore ConnectingEvent #1041
- slack: use blocks not attachments #1048
- sshchat: Update vendor shazow/ssh-chat #1029
- telegram: added markdownv2 mode for telegram #1037
- whatsapp: Implement basic reconnect (whatsapp). Fixes #987 #1003

## Bugfix

- discord: Fix webhook permission checks sometimes failing #1043
- discord: Fix #1027: warning when handling inbound webhooks #1044
- discord: Fix duplicate separator on empty description/url (discord) #1035
- matrix: Fix issue with underscores in links #999
- slack: Fix #1039: messages sent to Slack being synced back #1046
- telegram: Make avatars download work with mediaserverdownload (telegram). Fixes #920

This release couldn't exist without the following contributors:
@qaisjp, @jakubgs, @burner1024, @notpushkin, @MartijnBraam, @42wim

# v1.16.5

- Fix version bump

# v1.16.4

## New features

- whatsapp: Add support for WhatsApp media (jpeg/png/gif) bridging (#974)
- telegram: Add QuoteLengthLimit option (telegram) fixes #963 (#985)
- telegram: Add DisableWebPagePreview option (telegram). Closes #980 (#994)

## Enhancements

- general: update dependencies
- tengo: update to tengo v2
- general: Add Docker Compose configuration (#990)

## Bugfix

- general: Fail with message instead of panic. #988 (#991)
- telegram: Add extra mimetypes to docker image. Fixes #969
- discord: Fix channel ID problem with multiple gateways (discord). Fixes #953 (#977)
- discord: Show file comment in webhook if normal message is empty (discord). Fixes #962 (#995)
- matrix: Fix parsing issues - Disable smartypants in markdown parser. Fixes #989, #983 (#993)
- sshchat: Fix duplicated messages (sshchat). Fixes #950 (#996)

This release couldn't exist without the following contributors:
@jwflory, @42wim, @pbek, @Humorhenker, @c0ncord2, @glazzara

# v1.16.3

## Bugfix

- slack: Fix issues with ratelimiting #959
- mattermost: Fix bug when using webhookURL and login/token together #960

# v1.16.2

## New features

- keybase: Add support for receiving attachments (keybase) (#923)

## Enhancements

- general: Switch to new emoji library kyokomi/emoji (#948)
- general: Update markdown parsing library to github.com/gomarkdown/markdown (#944)
- ssh-chat: Update shazow/ssh-chat dependency (#947)

## Bugfix

- slack: Fix issues with the slack block kit API #937 (#943).

This release couldn't exist without the following contributors:
@42wim, @bmpickford, @goncalor

# v1.16.1

## New features

- rocketchat: add token support #892
- matrix: Add support for uploading application/x and audio/x (matrix). #929

## Enhancements

- general: Do configuration validation on start-up. Fixes #888
- general: updated vendored libraries (discord/whatsapp) #932
- discord: user typing messages #914
- slack: Convert slack bold/strike to correct markdown (slack). Fixes #918

## Bugfix

- discord: fix Failed to fetch information for members message. #894
- discord: remove obsolete file upload links (discord). #931
- slack: suppress unhandled HelloEvent message #913
- mattermost: Fix panic on WebhookURL only setting (mattermost). #917
- matrix: fix corrupted links between slack and matrix #924

This release couldn't exist without the following contributors:
@qaisjp, @hramrach, @42wim

# v1.16.0

## New features

- keybase: new protocol added. Add initial Keybase Chat support #877 Thanks to @hyperobject
- discord: Support webhook files in discord #872

## Enhancements

- general: update dependencies

## Bugfix

- discord: Underscores from Discord don't arrive correctly #864
- xmpp: Fix possible panic at startup of the XMPP bridge #869
- mattermost: Make getChannelIdTeam behave like GetChannelId for groups (mattermost) #873

This release couldn't exist without the following contributors:
@hyperobject, @42wim, @bucko909, @MOZGIII

# v1.15.1

## New features

- discord: Support webhook message deletions (discord) (#853)

## Enhancements

- discord: Support bulk deletions #851
- discord: Support channels in categories #863 (use category/channel. See matterbridge.toml.sample for more info)
- mattermost: Add an option to skip the Mattermost server version check #849

## Bugfix

- xmpp: fix segfault when disconnected/reconnected #856
- telegram: fix panic in handleEntities #858

This release couldn't exist without the following contributors:
@42wim, @qaisjp, @joohoi

# v1.15.0

## New features

- Add scripting (tengo) support for every outgoing message (#806)
  See https://github.com/42wim/matterbridge/wiki/Settings#tengo and
  https://github.com/42wim/matterbridge/wiki/Settings#outmessage for more information
- Add tengo support to RemoteNickFormat (#793)
  See https://github.com/42wim/matterbridge/wiki/Settings#remotenickformat-2
  - Deprecated `Message` under `[tengo]` to `InMessage`

## Enhancements

- general: Forward only user-typing messages if supported by protocol (#832)
- general: updated wiki with all possible settings: https://github.com/42wim/matterbridge/wiki/Settings
- tengo: Add msg event to tengo
- xmpp: Verify TLS against JID domain, not the host. (xmpp) (#834)
- xmpp: Allow messages with timestamp (xmpp). Fixes #835 (#847)
- irc: Add verbose IRC joins/parts (ident@host) (#805)
  See https://github.com/42wim/matterbridge/wiki/Settings#verbosejoinpart
- rocketchat: Add useraction support (rocketchat). Closes #772 (#794)

## Bugfix

- slack: Fix regression in autojoining with legacy tokens (slack). Fixes #651 (#848)
- xmpp: Revert xmpp to orig behaviour. Closes #844
- whatsapp: Update github.com/Rhymen/go-whatsapp vendor. Fixes #843
- mattermost: Update channels of all teams (mattermost)

This release couldn't exist without the following contributors:
@42wim, @Helcaraxan, @chotaire, @qaisjp, @dajohi, @kousu

# v1.14.4

## Bugfix

- mattermost: Add Id to EditMessage (mattermost). Fixes #802
- mattermost: Fix panic on nil message.Post (mattermost). Fixes #804
- mattermost: Handle unthreaded messages (mattermost). Fixes #803
- mattermost: Use paging in initUser and UpdateUsers (mattermost)
- slack: Add lacking clean-up in Slack synchronisation (#811)
- slack: Disable user lookups on delete messages (slack) (#812)

# v1.14.3

## Bugfix

- irc: Fix deadlock on reconnect (irc). Closes #757

# v1.14.2

## Bugfix

- general: Update tengo vendor and load the stdlib. Fixes #789 (#792)
- rocketchat: Look up #channel too (rocketchat). Fix #773 (#775)
- slack: Ignore messagereplied and hidden messages (slack). Fixes #709 (#779)
- telegram: Handle nil message (telegram). Fixes #777
- irc: Use default nick if none specified (irc). Fixes #785
- irc: Return when not connected and drop a message (irc). Fixes #786
- irc: Revert fix for #722 (Support quits from irc correctly). Closes #781

## Contributors

This release couldn't exist without the following contributors:
@42wim, @Helcaraxan, @dajohi

# v1.14.1

## Bugfix

- slack: Fix crash double unlock (slack) (#771)

# v1.14.0

## Breaking

- zulip: Need to specify /topic:mytopic for channel configuration (zulip). (#751)

## New features

- whatsapp: new protocol added. Add initial WhatsApp support (#711) Thanks to @KrzysztofMadejski
- facebook messenger: new protocol via matterbridge api. See https://github.com/VictorNine/fbridge/ for more information.
- general: Add scripting (tengo) support for every incoming message (#731). See `TengoModifyMessage`
- general: Allow regexs in ignoreNicks. Closes #690 (#720)
- general: Support rewriting messages from relaybots using ExtractNicks. Fixes #466 (#730). See `ExtractNicks` in matterbridge.toml.sample
- general: refactor Make all loggers derive from non-default instance (#728). Thanks to @Helcaraxan
- rocketchat: add support for the rocketchat API. Sending to rocketchat now supports uploading of files, editing and deleting of messages.
- discord: Support join/leaves from discord. Closes #654 (#721)
- discord: Allow sending discriminator with Discord username (#726). See `UseDiscriminator` in matterbridge.toml.sample
- slack: Add extra debug option (slack). See `Debug` in the slack section in matterbridge.toml.sample
- telegram: Add support for URL in messageEntities (telegram). Fixes #735 (#736)
- telegram: Add MediaConvertWebPToPNG option (telegram). (#741). See `MediaConvertWebPToPNG` in matterbridge.toml.sample

## Enhancements

- general: Fail gracefully on incorrect human input. Fixes #739 (#740)
- matrix: Detect html nicks in RemoteNickFormat (matrix). Fixes #696 (#719)
- matrix: Send notices on join/parts (matrix). Fixes #712 (#716)

## Bugfix

- general: Handle file upload/download only once for each message (#742)
- zulip: Fix error handling on bad event queue id (zulip). Closes #694
- zulip: Keep reconnecting until succeed (zulip) (#737)
- irc: add support for (older) unrealircd versions. #708
- irc: Support quits from irc correctly. Fixes #722 (#724)
- matrix: Send username when uploading video/images (matrix). Fixes #715 (#717)
- matrix: Trim <p> and </p> tags (matrix). Closes #686 (#753)
- slack: Hint at thread replies when messages are unthreaded (slack) (#684)
- slack: Fix race-condition in populateUser() (#767)
- xmpp: Do not send topic changes on connect (xmpp). Fixes #732 (#733)
- telegram: Fix regression in HTML handling (telegram). Closes #734
- discord: Do not relay any bot messages (discord) (#743)
- rocketchat: Do not send duplicate messages (rocketchat). Fixes #745 (#752)

## Contributors

This release couldn't exist without the following contributors:
@Helcaraxan, @KrzysztofMadejski, @AJolly, @DeclanHoare

# v1.13.1

This release fixes go modules issues because of https://github.com/labstack/echo/issues/1272

## Bugfix

- general: fixes Unable to build 1.13.0 #698
- api: move to labstack/echo/v4 fixes #698

# v1.13.0

## New features

- general: refactors of telegram, irc, mattermost, matrix, discord, sshchat bridges and the gateway.
- irc: Add option to send RAW commands after connection (irc) #490. See `RunCommands` in matterbridge.toml.sample
- mattermost: 3.x support dropped
- mattermost: Add support for mattermost threading (#627)
- slack: Sync channel topics between Slack bridges #585. See `SyncTopic` in matterbridge.toml.sample
- matrix: Add support for markdown to HTML conversion (matrix). Closes #663 (#670)
- discord: Improve error reporting on failure to join Discord. Fixes #672 (#680)
- discord: Use only one webhook if possible (discord) (#681)
- discord: Allow to bridge non-bot Discord users (discord) (#689) If you prefix a token with `User ` it'll treat is as a user token.

## Bugfix

- slack: Try downloading files again if slack is too slow (slack). Closes #655 (#656)
- slack: Ignore LatencyReport event (slack)
- slack: Fix #668 strip lang in code fences sent to Slack (#673)
- sshchat: Fix sshchat connection logic (#661)
- sshchat: set quiet mode to filter joins/quits
- sshchat: Trim newlines in the end of relayed messages
- sshchat: fix media links
- sshchat: do not relay "Rate limiting is in effect" message
- mattermost: Fail if channel starts with hashtag (mattermost). Closes #625
- discord: Add file comment to webhook messages (discord). Fixes #358
- matrix: Fix displaying usernames for plain text clients. (matrix) (#685)
- irc: Fix possible data race (irc). Closes #693
- irc: Handle servers without MOTD (irc). Closes #692

# v1.12.3

## Bugfix

- slack: Fix bot (legacy token) messages not being send. Closes #571
- slack: Populate user on channel join (slack) (#644)
- slack: Add wait option for populateUsers/Channels (slack) Fixes #579 (#653)

# v1.12.2

## Bugfix

- irc: Fix multiple channel join regression. Closes #639
- slack: Make slack-legacy change less restrictive (#626)

# v1.12.1

## Bugfix

- discord: fix regression on server ID connection #619 #617
- discord: Limit discord username via webhook to 32 chars
- slack: Make sure threaded files stay in thread (slack). Fixes #590
- slack: Do not post empty messages (slack). Fixes #574
- slack: Handle deleted/edited thread starting messages (slack). Fixes #600 (#605)
- irc: Rework connection logic (irc)
- irc: Fix Nickserv logic (irc) #602

# v1.12.0

## Breaking changes

The slack bridge has been split in a `slack-legacy` and `slack` bridge.
If you're still using `legacy tokens` and want to keep using them you'll have to rename `slack` to `slack-legacy` in your configuration. See [wiki](<https://github.com/42wim/matterbridge/wiki/Section-Slack-(basic)#legacy-configuration>) for more information.

To migrate to the new bot-token based setup you can follow the instructions [here](https://github.com/42wim/matterbridge/wiki/Slack-bot-setup).

Slack legacy tokens may be deprecated by Slack at short notice, so it is STRONGLY recommended to use a proper bot-token instead.

## New features

- general: New {GATEWAY} variable for `RemoteNickFormat` #501. See `RemoteNickFormat` in matterbridge.toml.sample.
- general: New {CHANNEL} variable for `RemoteNickFormat` #515. See `RemoteNickFormat` in matterbridge.toml.sample.
- general: Remove hyphens when auto-loading envvars from viper config #545
- discord: You can mention discord-users from other bridges.
- slack: Preserve threading between Slack instances #529. See `PreserveThreading` in matterbridge.toml.sample.
- slack: Add ability to show when user is typing across Slack bridges #559
- slack: Add rate-limiting
- mattermost: Add support for mattermost [matterbridge plugin](https://github.com/matterbridge/mattermost-plugin)
- api: Respond with message on connect. #550
- api: Add a health endpoint to API #554

## Bugfix

- slack: Refactoring and making it better.
- slack: Restore file comments coming from Slack. #583
- irc: Fix IRC line splitting. #587
- mattermost: Fix cookie and personal token behaviour. #530
- mattermost: Check for expiring sessions and reconnect.

## Contributors

This release couldn't exist without the following contributors:
@jheiselman, @NikkyAI, @dajohi, @NetwideRogue, @patcon and @Helcaraxan

Special thanks to @Helcaraxan and @patcon for their work on improving/refactoring slack.

# v1.11.3

## Bugfix

- mattermost: fix panic when using webhooks #491
- slack: fix issues regarding API changes and lots of channels #489
- irc: fix rejoin on kick problem #488

# v1.11.2

## Bugfix

- slack: fix slack API changes regarding to files/images

# v1.11.1

## New features

- slack: Add support for slack channels by ID. Closes #436
- discord: Clip too long messages sent to discord (discord). Closes #440

## Bugfix

- general: fix possible panic on downloads that are too big #448
- general: Fix avatar uploads to work with MediaDownloadPath. Closes #454
- discord: allow receiving of topic changes/channel leave/joins from other bridges through the webhook
- discord: Add a space before url in file uploads (discord). Closes #461
- discord: Skip empty messages being sent with the webhook (discord). #469
- mattermost: Use nickname instead of username if defined (mattermost). Closes #452
- irc: Stop numbers being stripped after non-color control codes (irc) (#465)
- slack: Use UserID to look for avatar instead of username (slack). Closes #472

# v1.11.0

## New features

- general: Add config option MediaDownloadPath (#443). See `MediaDownloadPath` in matterbridge.toml.sample
- general: Add MediaDownloadBlacklist option. Closes #442. See `MediaDownloadBlacklist` in matterbridge.toml.sample
- xmpp: Add channel password support for XMPP (#451)
- xmpp: Add message correction support for XMPP (#437)
- telegram: Add support for MessageFormat=htmlnick (telegram). #444
- mattermost: Add support for mattermost 5.x

## Enhancements

- slack: Add Title from attachment slack message (#446)
- irc: Prevent white or black color codes (irc) (#434)

## Bugfix

- slack: Fix regexp in replaceMention (slack). (#435)
- irc: Reconnect on quit. (irc) See #431 (#445)
- sshchat: Ignore messages from ourself. (sshchat) Closes #439

# v1.10.1

## New features

- irc: Colorize username sent to IRC using its crc32 IEEE checksum (#423). See `ColorNicks` in matterbridge.toml.sample
- irc: Add support for CJK to/from utf-8 (irc). #400
- telegram: Add QuoteFormat option (telegram). Closes #413. See `QuoteFormat` in matterbridge.toml.sample
- xmpp: Send attached files to XMPP in different message with OOB data and without body (#421)

## Bugfix

- general: updated irc/xmpp/telegram libraries
- mattermost/slack/rocketchat: Fix iconurl regression. Closes #430
- mattermost/slack: Use uuid instead of userid. Fixes #429
- slack: Avatar spoofing from Slack to Discord with uppercase in nick doesn't work (#433)
- irc: Fix format string bug (irc) (#428)

# v1.10.0

## New features

- general: Add support for reloading all settings automatically after changing config except connection and gateway configuration. Closes #373
- zulip: New protocol support added (https://zulipchat.com)

## Enhancements

- general: Handle file comment better
- steam: Handle file uploads to mediaserver (steam)
- slack: Properly set Slack user who initiated slash command (#394)

## Bugfix

- general: Use only alphanumeric for file uploads to mediaserver. Closes #416
- general: Fix crash on invalid filenames
- general: Fix regression in ReplaceMessages and ReplaceNicks. Closes #407
- telegram: Fix possible nil when using channels (telegram). #410
- telegram: Fix panic (telegram). Closes #410
- telegram: Handle channel posts correctly
- mattermost: Update GetFileLinks to API_V4

# v1.9.1

## New features

- telegram: Add QuoteDisable option (telegram). Closes #399. See QuoteDisable in matterbridge.toml.sample

## Enhancements

- discord: Send mediaserver link to Discord in Webhook mode (discord) (#405)
- mattermost: Print list of valid team names when team not found (#390)
- slack: Strip markdown URLs with blank text (slack) (#392)

## Bugfix

- slack/mattermost: Make our callbackid more unique. Fixes issue with running multiple matterbridge on the same channel (slack,mattermost)
- telegram: fix newlines in multiline messages #399
- telegram: Revert #378

# v1.9.0 (the refactor release)

## New features

- general: better debug messages
- general: better support for environment variables override
- general: Ability to disable sending join/leave messages to other gateways. #382
- slack: Allow Slack @usergroups to be parsed as human-friendly names #379
- slack: Provide better context for shared posts from Slack<=>Slack enhancement #369
- telegram: Convert nicks automatically into HTML when MessageFormat is set to HTML #378
- irc: Add DebugLevel option

## Bugfix

- slack: Ignore restricted_action on channel join (slack). Closes #387
- slack: Add slack attachment support to matterhook
- slack: Update userlist on join (slack). Closes #372

# v1.8.0

## New features

- general: Send chat notification if media is too big to be re-uploaded to MediaServer. See #359
- general: Download (and upload) avatar images from mattermost and telegram when mediaserver is configured. Closes #362
- general: Add label support in RemoteNickFormat
- general: Prettier info/debug log output
- mattermost: Download files and reupload to supported bridges (mattermost). Closes #357
- slack: Add ShowTopicChange option. Allow/disable topic change messages (currently only from slack). Closes #353
- slack: Add support for file comments (slack). Closes #346
- telegram: Add comment to file upload from telegram. Show comments on all bridges. Closes #358
- telegram: Add markdown support (telegram). #355
- api: Give api access to whole config.Message (and events). Closes #374

## Bugfix

- discord: Check for a valid WebhookURL (discord). Closes #367
- discord: Fix role mention replace issues
- irc: Truncate messages sent to IRC based on byte count (#368)
- mattermost: Add file download urls also to mattermost webhooks #356
- telegram: Fix panic on nil messages (telegram). Closes #366
- telegram: Fix the UseInsecureURL text (telegram). Closes #184

# v1.7.1

## Bugfix

- telegram: Enable Long Polling for Telegram. Reduces bandwidth consumption. (#350)

# v1.7.0

## New features

- matrix: Add support for deleting messages from/to matrix (matrix). Closes #320
- xmpp: Ignore <subject> messages (xmpp). #272
- irc: Add twitch support (irc) to README / wiki

## Bugfix

- general: Change RemoteNickFormat replacement order. Closes #336
- general: Make edits/delete work for bridges that gets reused. Closes #342
- general: Lowercase irc channels in config. Closes #348
- matrix: Fix possible panics (matrix). Closes #333
- matrix: Add an extension to images without one (matrix). #331
- api: Obey the Gateway value from the json (api). Closes #344
- xmpp: Print only debug messages when specified (xmpp). Closes #345
- xmpp: Allow xmpp to receive the extra messages (file uploads) when text is empty. #295

# v1.6.3

## Bugfix

- slack: Fix connection issues
- slack: Add more debug messages
- irc: Convert received IRC channel names to lowercase. Fixes #329 (#330)

# v1.6.2

## Bugfix

- mattermost: Crashes while connecting to Mattermost (regression). Closes #327

# v1.6.1

## Bugfix

- general: Display of nicks not longer working (regression). Closes #323

# v1.6.0

## New features

- sshchat: New protocol support added (https://github.com/shazow/ssh-chat)
- general: Allow specifying maximum download size of media using MediaDownloadSize (slack,telegram,matrix)
- api: Add (simple, one listener) long-polling support (api). Closes #307
- telegram: Add support for forwarded messages. Closes #313
- telegram: Add support for Audio/Voice files (telegram). Closes #314
- irc: Add RejoinDelay option. Delay to rejoin after channel kick (irc). Closes #322

## Bugfix

- telegram: Also use HTML in edited messages (telegram). Closes #315
- matrix: Fix panic (matrix). Closes #316

# v1.5.1

## Bugfix

- irc: Fix irc ACTION regression (irc). Closes #306
- irc: Split on UTF-8 for MessageSplit (irc). Closes #308

# v1.5.0

## New features

- general: remote mediaserver support. See MediaServerDownload and MediaServerUpload in matterbridge.toml.sample
  more information on https://github.com/42wim/matterbridge/wiki/Mediaserver-setup-%5Badvanced%5D
- general: Add support for ReplaceNicks using regexp to replace nicks. Closes #269 (see matterbridge.toml.sample)
- general: Add support for ReplaceMessages using regexp to replace messages. #269 (see matterbridge.toml.sample)
- irc: Add MessageSplit option to split messages on MessageLength (irc). Closes #281
- matrix: Add support for uploading images/video (matrix). Closes #302
- matrix: Add support for uploaded images/video (matrix)

## Bugfix

- telegram: Add webp extension to stickers if necessary (telegram)
- mattermost: Break when re-login fails (mattermost)

# v1.4.1

## Bugfix

- telegram: fix issue with uploading for images/documents/stickers
- slack: remove double messages sent to other bridges when uploading files
- irc: Fix strict user handling of girc (irc). Closes #298

# v1.4.0

## Breaking changes

- general: `[general]` settings don't override the specific bridge settings

## New features

- irc: Replace sorcix/irc and go-ircevent with girc, this should be give better reconnects
- steam: Add support for bridging to individual steam chats. (steam) (#294)
- telegram: Download files from telegram and reupload to supported bridges (telegram). #278
- slack: Add support to upload files to slack, from bridges with private urls like slack/mattermost/telegram. (slack)
- discord: Add support to upload files to discord, from bridges with private urls like slack/mattermost/telegram. (discord)
- general: Add systemd service file (#291)
- general: Add support for DEBUG=1 envvar to enable debug. Closes #283
- general: Add StripNick option, only allow alphanumerical nicks. Closes #285

## Bugfix

- gitter: Use room.URI instead of room.Name. (gitter) (#293)
- slack: Allow slack messages with variables (eg. @here) to be formatted correctly. (slack) (#288)
- slack: Resolve slack channel to human-readable name. (slack) (#282)
- slack: Use DisplayName instead of deprecated username (slack). Closes #276
- slack: Allowed Slack bridge to extract simpler link format. (#287)
- irc: Strip irc colors correct, strip also ctrl chars (irc)

# v1.3.1

## New features

- Support mattermost 4.3.0 and every other 4.x as api4 should be stable (mattermost)

## Bugfix

- Use bot username if specified (slack). Closes #273

# v1.3.0

## New features

- Relay slack_attachments from mattermost to slack (slack). Closes #260
- Add support for quoting previous message when replying (telegram). #237
- Add support for Quakenet auth (irc). Closes #263
- Download files (max size 1MB) from slack and reupload to mattermost (slack/mattermost). Closes #255

## Enhancements

- Backoff for 60 seconds when reconnecting too fast (irc) #267
- Use override username if specified (mattermost). #260

## Bugfix

- Try to not forward slack unfurls. Closes #266

# v1.2.0

## Breaking changes

- If you're running a discord bridge, update to this release before 16 october otherwise
  it will stop working. (see https://discordapp.com/developers/docs/reference)

## New features

- general: Add delete support. (actually delete the messages on bridges that support it)
  (mattermost,discord,gitter,slack,telegram)

## Bugfix

- Do not break messages on newline (slack). Closes #258
- Update telegram library
- Update discord library (supports v6 API now). Old API is deprecated on 16 October

# v1.1.2

## New features

- general: also build darwin binaries
- mattermost: add support for mattermost 4.2.x

## Bugfix

- mattermost: Send images when text is empty regression. (mattermost). Closes #254
- slack: also send the first messsage after connect. #252

# v1.1.1

## Bugfix

- mattermost: fix public links

# v1.1.0

## New features

- general: Add better editing support. (actually edit the messages on bridges that support it)
  (mattermost,discord,gitter,slack,telegram)
- mattermost: use API v4 (removes support for mattermost < 3.8)
- mattermost: add support for personal access tokens (since mattermost 4.1)
  Use `Token="yourtoken"` in mattermost config
  See https://docs.mattermost.com/developer/personal-access-tokens.html for more info
- matrix: Relay notices (matrix). Closes #243
- irc: Add a charset option. Closes #247

## Bugfix

- slack: Handle leave/join events (slack). Closes #246
- slack: Replace mentions from other bridges. (slack). Closes #233
- gitter: remove ZWSP after messages

# v1.0.1

## New features

- mattermost: add support for mattermost 4.1.x
- discord: allow a webhookURL per channel #239

# v1.0.0

## New features

- general: Add action support for slack,mattermost,irc,gitter,matrix,xmpp,discord. #199
- discord: Shows the username instead of the server nickname #234

# v1.0.0-rc1

## New features

- general: Add action support for slack,mattermost,irc,gitter,matrix,xmpp,discord. #199

## Bugfix

- general: Handle same account in multiple gateways better
- mattermost: ignore edited messages with reactions
- mattermost: Fix double posting of edited messages by using lru cache
- irc: update vendor

# v0.16.3

## Bugfix

- general: Fix in/out logic. Closes #224
- general: Fix message modification
- slack: Disable message from other bots when using webhooks (slack)
- mattermost: Return better error messages on mattermost connect

# v0.16.2

## New features

- general: binary builds against latest commit are now available on https://bintray.com/42wim/nightly/Matterbridge/_latestVersion

## Bugfix

- slack: fix loop introduced by relaying message of other bots #219
- slack: Suppress parent message when child message is received #218
- mattermost: fix regression when using webhookurl and webhookbindaddress #221

# v0.16.1

## New features

- slack: also relay messages of other bots #213
- mattermost: show also links if public links have not been enabled.

## Bugfix

- mattermost, slack: fix connecting logic #216

# v0.16.0

## Breaking Changes

- URL,UseAPI,BindAddress is deprecated. Your config has to be updated.
  - URL => WebhookURL
  - BindAddress => WebhookBindAddress
  - UseAPI => removed
    This change allows you to specify a WebhookURL and a token (slack,discord), so that
    messages will be sent with the webhook, but received via the token (API)
    If you have not specified WebhookURL and WebhookBindAddress the API (login or token)
    will be used automatically. (no need for UseAPI)

## New features

- mattermost: add support for mattermost 4.0
- steam: New protocol support added (http://store.steampowered.com/)
- discord: Support for embedded messages (sent by other bots)
  Shows title, description and URL of embedded messages (sent by other bots)
  To enable add `ShowEmbeds=true` to your discord config
- discord: `WebhookURL` posting support added (thanks @saury07) #204
  Discord API does not allow to change the name of the user posting, but webhooks does.

## Changes

- general: all :emoji: will be converted to unicode, providing consistent emojis across all bridges
- telegram: Add `UseInsecureURL` option for telegram (default false)
  WARNING! If enabled this will relay GIF/stickers/documents and other attachments as URLs
  Those URLs will contain your bot-token. This may not be what you want.
  For now there is no secure way to relay GIF/stickers/documents without seeing your token.

## Bugfix

- irc: detect charset and try to convert it to utf-8 before sending it to other bridges. #209 #210
- slack: Remove label from URLs (slack). #205
- slack: Relay <>& correctly to other bridges #215
- steam: Fix channel id bug in steam (channels are off by 0x18000000000000)
- general: various improvements
- general: samechannelgateway now relays messages correct again #207

# v0.16.0-rc2

## Breaking Changes

- URL,UseAPI,BindAddress is deprecated. Your config has to be updated.
  - URL => WebhookURL
  - BindAddress => WebhookBindAddress
  - UseAPI => removed
    This change allows you to specify a WebhookURL and a token (slack,discord), so that
    messages will be sent with the webhook, but received via the token (API)
    If you have not specified WebhookURL and WebhookBindAddress the API (login or token)
    will be used automatically. (no need for UseAPI)

## Bugfix since rc1

- steam: Fix channel id bug in steam (channels are off by 0x18000000000000)
- telegram: Add UseInsecureURL option for telegram (default false)
  WARNING! If enabled this will relay GIF/stickers/documents and other attachments as URLs
  Those URLs will contain your bot-token. This may not be what you want.
  For now there is no secure way to relay GIF/stickers/documents without seeing your token.
- irc: detect charset and try to convert it to utf-8 before sending it to other bridges. #209 #210
- general: various improvements

# v0.16.0-rc1

## Breaking Changes

- URL,UseAPI,BindAddress is deprecated. Your config has to be updated.
  - URL => WebhookURL
  - BindAddress => WebhookBindAddress
  - UseAPI => removed
    This change allows you to specify a WebhookURL and a token (slack,discord), so that
    messages will be sent with the webhook, but received via the token (API)
    If you have not specified WebhookURL and WebhookBindAddress the API (login or token)
    will be used automatically. (no need for UseAPI)

## New features

- steam: New protocol support added (http://store.steampowered.com/)
- discord: WebhookURL posting support added (thanks @saury07) #204
  Discord API does not allow to change the name of the user posting, but webhooks does.

## Bugfix

- general: samechannelgateway now relays messages correct again #207
- slack: Remove label from URLs (slack). #205

# v0.15.0

## New features

- general: add option IgnoreMessages for all protocols (see mattebridge.toml.sample)
  Messages matching these regexp will be ignored and not sent to other bridges
  e.g. IgnoreMessages="^~~ badword"
- telegram: add support for sticker/video/photo/document #184

## Changes

- api: add userid to each message #200

## Bugfix

- discord: fix crash in memberupdate #198
- mattermost: Fix incorrect behaviour of EditDisable (mattermost). Fixes #197
- irc: Do not relay join/part of ourselves (irc). Closes #190
- irc: make reconnections more robust. #153
- gitter: update library, fixes possible crash

# v0.14.0

## New features

- api: add token authentication
- mattermost: add support for mattermost 3.10.0

## Changes

- api: gateway name is added in JSON messages
- api: lowercase JSON keys
- api: channel name isn't needed in config #195

## Bugfix

- discord: Add hashtag to channelname (when translating from id) (discord)
- mattermost: Fix a panic. #186
- mattermost: use teamid cache if possible. Fixes a panic
- api: post valid json. #185
- api: allow reuse of api in different gateways. #189
- general: Fix utf-8 issues for {NOPINGNICK}. #193

# v0.13.0

## New features

- irc: Limit message length. `MessageLength=400`
  Maximum length of message sent to irc server. If it exceeds <message clipped> will be add to the message.
- irc: Add NOPINGNICK option.
  The string "{NOPINGNICK}" (case sensitive) will be replaced by the actual nick / username, but with a ZWSP inside the nick, so the irc user with the same nick won't get pinged.  
  See https://github.com/42wim/matterbridge/issues/175 for more information

## Bugfix

- slack: Fix sending to different channels on same account (slack). Closes #177
- telegram: Fix incorrect usernames being sent. Closes #181

# v0.12.1

## New features

- telegram: Add UseFirstName option (telegram). Closes #144
- matrix: Add NoHomeServerSuffix. Option to disable homeserver on username (matrix). Closes #160.

## Bugfix

- xmpp: Add Compatibility for Cisco Jabber (xmpp) (#166)
- irc: Fix JoinChannel argument to use IRC channel key (#172)
- discord: Fix possible crash on nil (discord)
- discord: Replace long ids in channel metions (discord). Fixes #174

# v0.12.0

## Changes

- general: edited messages are now being sent by default on discord/mattermost/telegram/slack. See "New Features"

## New features

- general: add support for edited messages.
  Add new keyword EditDisable (false/true), default false. Which means by default edited messages will be sent to other bridges.
  Add new keyword EditSuffix , default "". You can change this eg to "(edited)", this will be appended to every edit message.
- mattermost: support mattermost v3.9.x
- general: Add support for HTTP{S}\_PROXY env variables (#162)
- discord: Strip custom emoji metadata (discord). Closes #148

## Bugfix

- slack: Ignore error on private channel join (slack) Fixes #150
- mattermost: fix crash on reconnects when server is down. Closes #163
- irc: Relay messages starting with ! (irc). Closes #164

# v0.11.0

## New features

- general: reusing the same account on multiple gateways now also reuses the connection.
  This is particuarly useful for irc. See #87
- general: the Name is now REQUIRED and needs to be UNIQUE for each gateway configuration
- telegram: Support edited messages (telegram). See #141
- mattermost: Add support for showing/hiding join/leave messages from mattermost. Closes #147
- mattermost: Reconnect on session removal/timeout (mattermost)
- mattermost: Support mattermost v3.8.x
- irc: Rejoin channel when kicked (irc).

## Bugfix

- mattermost: Remove space after nick (mattermost). Closes #142
- mattermost: Modify iconurl correctly (mattermost).
- irc: Fix join/leave regression (irc)

# v0.10.3

## Bugfix

- slack: Allow bot tokens for now without warning (slack). Closes #140 (fixes user_is_bot message on channel join)

# v0.10.2

## New features

- general: gops agent added. Allows for more debugging. See #134
- general: toml inline table support added for config file

## Bugfix

- all: vendored libs updated

## Changes

- general: add more informative messages on startup

# v0.10.1

## Bugfix

- gitter: Fix sending messages on new channel join.

# v0.10.0

## New features

- matrix: New protocol support added (https://matrix.org)
- mattermost: works with mattermost release v3.7.0
- discord: Replace role ids in mentions to role names (discord). Closes #133

## Bugfix

- mattermost: Add ReadTimeout to close lingering connections (mattermost). See #125
- gitter: Join rooms not already joined by the bot (gitter). See #135
- general: Fail when bridge is unable to join a channel (general)

## Changes

- telegram: Do not use HTML parsemode by default. Set `MessageFormat="HTML"` to use it. Closes #126

# v0.9.3

## New features

- API: rest interface to read / post messages (see API section in matterbridge.toml.sample)

## Bugfix

- slack: fix receiving messages from private channels #118
- slack: fix echo when using webhooks #119
- mattermost: reconnecting should work better now
- irc: keeps reconnecting (every 60 seconds) now after ping timeout/disconnects.

# v0.9.2

## New features

- slack: support private channels #118

## Bugfix

- general: make ignorenicks work again #115
- telegram: fix receiving from channels and groups #112
- telegram: use html for username
- telegram: use `unknown` as username when username is not visible.
- irc: update vendor (fixes some crashes) #117
- xmpp: fix tls by setting ServerName #114

# v0.9.1

## New features

- Rocket.Chat: New protocol support added (https://rocket.chat)
- irc: add channel key support #27 (see matterbrige.toml.sample for example)
- xmpp: add SkipTLSVerify #106

## Bugfix

- general: Exit when a bridge fails to start
- mattermost: Check errors only on first connect. Keep retrying after first connection succeeds. #95
- telegram: fix missing username #102
- slack: do not use API functions in webhook (slack) #110

# v0.9.0

## New features

- Telegram: New protocol support added (https://telegram.org)
- Hipchat: Add sample config to connect to hipchat via xmpp
- discord: add "Bot " tag to discord tokens automatically
- slack: Add support for dynamic Iconurl #43
- general: Add `gateway.inout` config option for bidirectional bridges #85
- general: Add `[general]` section so that `RemoteNickFormat` can be set globally

## Bugfix

- general: when using samechannelgateway NickFormat get doubled by the NICK #77
- general: fix ShowJoinPart for messages from irc bridge #72
- gitter: fix high cpu usage #89
- irc: fix !users command #78
- xmpp: fix keepalive
- xmpp: do not relay delayed/empty messages
- slack: Replace id-mentions to usernames #86
- mattermost: fix public links not working (API changes)

# v0.8.1

## Bugfix

- general: when using samechannelgateway NickFormat get doubled by the NICK #77
- irc: fix !users command #78

# v0.8.0

Release because of breaking mattermost API changes

## New features

- Supports mattermost v3.5.0

# v0.7.1

## Bugfix

- general: when using samechannelgateway NickFormat get doubled by the NICK #77
- irc: fix !users command #78

# v0.7.0

## Breaking config changes from 0.6 to 0.7

Matterbridge now uses TOML configuration (https://github.com/toml-lang/toml)
See matterbridge.toml.sample for an example

## New features

### General

- Allow for bridging the same type of bridge, which means you can eg bridge between multiple mattermosts.
- The bridge is now actually a gateway which has support multiple in and out bridges. (and supports multiple gateways).
- Discord support added. See matterbridge.toml.sample for more information.
- Samechannelgateway support added, easier configuration for 1:1 mapping of protocols with same channel names. #35
- Support for override from environment variables. #50
- Better debugging output.
- discord: New protocol support added. (http://www.discordapp.com)
- mattermost: Support attachments.
- irc: Strip colors. #33
- irc: Anti-flooding support. #40
- irc: Forward channel notices.

## Bugfix

- irc: Split newlines. #37
- irc: Only respond to nick related notices from nickserv.
- irc: Ignore queries send to the bot.
- irc: Ignore messages from ourself.
- irc: Only output the "users on irc information" when asked with "!users".
- irc: Actually wait until connection is complete before saying it is.
- mattermost: Fix mattermost channel joins.
- mattermost: Drop messages not from our team.
- slack: Do not panic on non-existing channels.
- general: Exit when a bridge fails to start.

# v0.6.1

## New features

- Slack support added. See matterbridge.conf.sample for more information

## Bugfix

- Fix 100% CPU bug on incorrect closed connections

# v0.6.0-beta2

## New features

- Gitter support added. See matterbridge.conf.sample for more information

# v0.6.0-beta1

## Breaking changes from 0.5 to 0.6

### commandline

- -plus switch deprecated. Use `Plus=true` or `Plus` in `[general]` section

### IRC section

- `Enabled` added (default false)  
  Add `Enabled=true` or `Enabled` to the `[IRC]` section if you want to enable the IRC bridge

### Mattermost section

- `Enabled` added (default false)  
  Add `Enabled=true` or `Enabled` to the `[mattermost]` section if you want to enable the mattermost bridge

### General section

- Use `Plus=true` or `Plus` in `[general]` section to enable the API version of matterbridge

## New features

- Matterbridge now bridges between any specified protocol (not only mattermost anymore)
- XMPP support added. See matterbridge.conf.sample for more information
- RemoteNickFormat {BRIDGE} variable added  
  You can now add the originating bridge to `RemoteNickFormat`  
  eg `RemoteNickFormat="[{BRIDGE}] <{NICK}> "`

# v0.5.0

## Breaking changes from 0.4 to 0.5 for matterbridge (webhooks version)

### IRC section

#### Server

Port removed, added to server

```
server="irc.freenode.net"
port=6667
```

changed to

```
server="irc.freenode.net:6667"
```

#### Channel

Removed see Channels section below

#### UseSlackCircumfix=true

Removed, can be done by using `RemoteNickFormat="<{NICK}> "`

### Mattermost section

#### BindAddress

Port removed, added to BindAddress

```
BindAddress="0.0.0.0"
port=9999
```

changed to

```
BindAddress="0.0.0.0:9999"
```

#### Token

Removed

### Channels section

```
[Token "outgoingwebhooktoken1"]
IRCChannel="#off-topic"
MMChannel="off-topic"
```

changed to

```
[Channel "channelnameofchoice"]
IRC="#off-topic"
Mattermost="off-topic"
```
