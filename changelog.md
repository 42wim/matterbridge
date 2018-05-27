# v1.10.1
## New features
* irc: Colorize username sent to IRC using its crc32 IEEE checksum (#423). See `ColorNicks` in matterbridge.toml.sample
* irc: Add support for CJK to/from utf-8 (irc). #400
* telegram: Add QuoteFormat option (telegram). Closes #413. See `QuoteFormat` in matterbridge.toml.sample
* xmpp: Send attached files to XMPP in different message with OOB data and without body (#421)

## Bugfix
* general: updated irc/xmpp/telegram libraries
* mattermost/slack/rocketchat: Fix iconurl regression. Closes #430
* mattermost/slack: Use uuid instead of userid. Fixes #429
* slack: Avatar spoofing from Slack to Discord with uppercase in nick doesn't work (#433)
* irc: Fix format string bug (irc) (#428)

# v1.10.0
## New features
* general: Add support for reloading all settings automatically after changing config except connection and gateway configuration. Closes #373
* zulip: New protocol support added (https://zulipchat.com)

## Enhancements
* general: Handle file comment better
* steam: Handle file uploads to mediaserver (steam)
* slack: Properly set Slack user who initiated slash command (#394)

## Bugfix
* general: Use only alphanumeric for file uploads to mediaserver. Closes #416
* general: Fix crash on invalid filenames
* general: Fix regression in ReplaceMessages and ReplaceNicks. Closes #407
* telegram: Fix possible nil when using channels (telegram). #410
* telegram: Fix panic (telegram). Closes #410
* telegram: Handle channel posts correctly
* mattermost: Update GetFileLinks to API_V4

# v1.9.1
## New features
* telegram: Add QuoteDisable option (telegram). Closes #399. See QuoteDisable in matterbridge.toml.sample
## Enhancements
* discord: Send mediaserver link to Discord in Webhook mode (discord) (#405)
* mattermost: Print list of valid team names when team not found (#390)
* slack: Strip markdown URLs with blank text (slack) (#392)
## Bugfix
* slack/mattermost: Make our callbackid more unique. Fixes issue with running multiple matterbridge on the same channel (slack,mattermost)
* telegram: fix newlines in multiline messages #399
* telegram: Revert #378

# v1.9.0 (the refactor release)
## New features
* general: better debug messages
* general: better support for environment variables override
* general: Ability to disable sending join/leave messages to other gateways. #382
* slack: Allow Slack @usergroups to be parsed as human-friendly names #379
* slack: Provide better context for shared posts from Slack<=>Slack enhancement #369
* telegram: Convert nicks automatically into HTML when MessageFormat is set to HTML #378
* irc: Add DebugLevel option 

## Bugfix
* slack: Ignore restricted_action on channel join (slack). Closes #387
* slack: Add slack attachment support to matterhook
* slack: Update userlist on join (slack). Closes #372

# v1.8.0
## New features
* general: Send chat notification if media is too big to be re-uploaded to MediaServer. See #359
* general: Download (and upload) avatar images from mattermost and telegram when mediaserver is configured. Closes #362
* general: Add label support in RemoteNickFormat
* general: Prettier info/debug log output
* mattermost: Download files and reupload to supported bridges (mattermost). Closes #357
* slack: Add ShowTopicChange option. Allow/disable topic change messages (currently only from slack). Closes #353
* slack: Add support for file comments (slack). Closes #346
* telegram: Add comment to file upload from telegram. Show comments on all bridges. Closes #358
* telegram: Add markdown support (telegram). #355
* api: Give api access to whole config.Message (and events). Closes #374

## Bugfix
* discord: Check for a valid WebhookURL (discord). Closes #367
* discord: Fix role mention replace issues
* irc: Truncate messages sent to IRC based on byte count (#368)
* mattermost: Add file download urls also to mattermost webhooks #356
* telegram: Fix panic on nil messages (telegram). Closes #366
* telegram: Fix the UseInsecureURL text (telegram). Closes #184

# v1.7.1
## Bugfix
* telegram: Enable Long Polling for Telegram. Reduces bandwidth consumption. (#350)

# v1.7.0
## New features
* matrix: Add support for deleting messages from/to matrix (matrix). Closes #320
* xmpp: Ignore <subject> messages (xmpp). #272
* irc: Add twitch support (irc) to README / wiki

## Bugfix
* general: Change RemoteNickFormat replacement order. Closes #336
* general: Make edits/delete work for bridges that gets reused. Closes #342
* general: Lowercase irc channels in config. Closes #348
* matrix: Fix possible panics (matrix). Closes #333
* matrix: Add an extension to images without one (matrix). #331
* api: Obey the Gateway value from the json (api). Closes #344
* xmpp: Print only debug messages when specified (xmpp). Closes #345
* xmpp: Allow xmpp to receive the extra messages (file uploads) when text is empty. #295

# v1.6.3
## Bugfix
* slack: Fix connection issues
* slack: Add more debug messages
* irc: Convert received IRC channel names to lowercase. Fixes #329 (#330)

# v1.6.2
## Bugfix
* mattermost: Crashes while connecting to Mattermost (regression). Closes #327

# v1.6.1
## Bugfix
* general: Display of nicks not longer working (regression). Closes #323

# v1.6.0
## New features
* sshchat: New protocol support added (https://github.com/shazow/ssh-chat)
* general: Allow specifying maximum download size of media using MediaDownloadSize (slack,telegram,matrix)
* api: Add (simple, one listener) long-polling support (api). Closes #307
* telegram: Add support for forwarded messages. Closes #313
* telegram: Add support for Audio/Voice files (telegram). Closes #314
* irc: Add RejoinDelay option. Delay to rejoin after channel kick (irc). Closes #322

## Bugfix
* telegram: Also use HTML in edited messages (telegram). Closes #315
* matrix: Fix panic (matrix). Closes #316

# v1.5.1

## Bugfix
* irc: Fix irc ACTION regression (irc). Closes #306
* irc: Split on UTF-8 for MessageSplit (irc). Closes #308

# v1.5.0
## New features
* general: remote mediaserver support. See MediaServerDownload and MediaServerUpload in matterbridge.toml.sample
  more information on https://github.com/42wim/matterbridge/wiki/Mediaserver-setup-%5Badvanced%5D
* general: Add support for ReplaceNicks using regexp to replace nicks. Closes #269 (see matterbridge.toml.sample)
* general: Add support for ReplaceMessages using regexp to replace messages. #269 (see matterbridge.toml.sample)
* irc: Add MessageSplit option to split messages on MessageLength (irc). Closes #281
* matrix: Add support for uploading images/video (matrix). Closes #302
* matrix: Add support for uploaded images/video (matrix) 

## Bugfix
* telegram: Add webp extension to stickers if necessary (telegram)
* mattermost: Break when re-login fails (mattermost)

# v1.4.1
## Bugfix
* telegram: fix issue with uploading for images/documents/stickers
* slack: remove double messages sent to other bridges when uploading files
* irc: Fix strict user handling of girc (irc). Closes #298 

# v1.4.0
## Breaking changes
* general: `[general]` settings don't override the specific bridge settings

## New features
* irc: Replace sorcix/irc and go-ircevent with girc, this should be give better reconnects
* steam: Add support for bridging to individual steam chats. (steam) (#294)
* telegram: Download files from telegram and reupload to supported bridges (telegram). #278
* slack: Add support to upload files to slack, from bridges with private urls like slack/mattermost/telegram. (slack)
* discord: Add support to upload files to discord, from bridges with private urls like slack/mattermost/telegram. (discord)
* general: Add systemd service file (#291)
* general: Add support for DEBUG=1 envvar to enable debug. Closes #283
* general: Add StripNick option, only allow alphanumerical nicks. Closes #285

## Bugfix
* gitter: Use room.URI instead of room.Name. (gitter) (#293)
* slack: Allow slack messages with variables (eg. @here) to be formatted correctly. (slack) (#288)
* slack: Resolve slack channel to human-readable name. (slack) (#282)
* slack: Use DisplayName instead of deprecated username (slack). Closes #276
* slack: Allowed Slack bridge to extract simpler link format. (#287)
* irc: Strip irc colors correct, strip also ctrl chars (irc)

# v1.3.1
## New features
* Support mattermost 4.3.0 and every other 4.x as api4 should be stable (mattermost)
## Bugfix
* Use bot username if specified (slack). Closes #273

# v1.3.0
## New features
* Relay slack_attachments from mattermost to slack (slack). Closes #260
* Add support for quoting previous message when replying (telegram). #237
* Add support for Quakenet auth (irc). Closes #263
* Download files (max size 1MB) from slack and reupload to mattermost (slack/mattermost). Closes #255

## Enhancements
* Backoff for 60 seconds when reconnecting too fast (irc) #267
* Use override username if specified (mattermost). #260

## Bugfix
* Try to not forward slack unfurls. Closes #266

# v1.2.0
## Breaking changes
* If you're running a discord bridge, update to this release before 16 october otherwise
it will stop working. (see https://discordapp.com/developers/docs/reference)

## New features
* general: Add delete support. (actually delete the messages on bridges that support it)
    (mattermost,discord,gitter,slack,telegram)

## Bugfix
* Do not break messages on newline (slack). Closes #258 
* Update telegram library
* Update discord library (supports v6 API now). Old API is deprecated on 16 October

# v1.1.2
## New features
* general: also build darwin binaries
* mattermost: add support for mattermost 4.2.x

## Bugfix 
* mattermost: Send images when text is empty regression. (mattermost). Closes #254
* slack: also send the first messsage after connect. #252

# v1.1.1
## Bugfix
* mattermost: fix public links

# v1.1.0
## New features
* general: Add better editing support. (actually edit the messages on bridges that support it)
	(mattermost,discord,gitter,slack,telegram)
* mattermost: use API v4 (removes support for mattermost < 3.8)
* mattermost: add support for personal access tokens (since mattermost 4.1)
	Use ```Token="yourtoken"``` in mattermost config
	See https://docs.mattermost.com/developer/personal-access-tokens.html for more info
* matrix: Relay notices (matrix). Closes #243
* irc: Add a charset option. Closes #247

## Bugfix
* slack: Handle leave/join events (slack). Closes #246
* slack: Replace mentions from other bridges. (slack). Closes #233
* gitter: remove ZWSP after messages

# v1.0.1
## New features
* mattermost: add support for mattermost 4.1.x
* discord: allow a webhookURL per channel #239

# v1.0.0
## New features
* general: Add action support for slack,mattermost,irc,gitter,matrix,xmpp,discord. #199
* discord: Shows the username instead of the server nickname #234

# v1.0.0-rc1
## New features
* general: Add action support for slack,mattermost,irc,gitter,matrix,xmpp,discord. #199

## Bugfix
* general: Handle same account in multiple gateways better
* mattermost: ignore edited messages with reactions
* mattermost: Fix double posting of edited messages by using lru cache
* irc: update vendor

# v0.16.3
## Bugfix
* general: Fix in/out logic. Closes #224 
* general: Fix message modification
* slack: Disable message from other bots when using webhooks (slack)
* mattermost: Return better error messages on mattermost connect

# v0.16.2
## New features
* general: binary builds against latest commit are now available on https://bintray.com/42wim/nightly/Matterbridge/_latestVersion

## Bugfix
* slack: fix loop introduced by relaying message of other bots #219
* slack: Suppress parent message when child message is received #218
* mattermost: fix regression when using webhookurl and webhookbindaddress #221

# v0.16.1
## New features
* slack: also relay messages of other bots #213
* mattermost: show also links if public links have not been enabled.

## Bugfix
* mattermost, slack: fix connecting logic #216

# v0.16.0
## Breaking Changes
* URL,UseAPI,BindAddress is deprecated. Your config has to be updated.
  * URL => WebhookURL
  * BindAddress => WebhookBindAddress
  * UseAPI => removed 
  This change allows you to specify a WebhookURL and a token (slack,discord), so that
  messages will be sent with the webhook, but received via the token (API)
  If you have not specified WebhookURL and WebhookBindAddress the API (login or token) 
  will be used automatically. (no need for UseAPI)

## New features
* mattermost: add support for mattermost 4.0
* steam: New protocol support added (http://store.steampowered.com/)
* discord: Support for embedded messages (sent by other bots)
  Shows title, description and URL of embedded messages (sent by other bots)
  To enable add ```ShowEmbeds=true``` to your discord config 
* discord: ```WebhookURL``` posting support added (thanks @saury07) #204
  Discord API does not allow to change the name of the user posting, but webhooks does.

## Changes
* general: all :emoji: will be converted to unicode, providing consistent emojis across all bridges
* telegram: Add ```UseInsecureURL``` option for telegram (default false)
  WARNING! If enabled this will relay GIF/stickers/documents and other attachments as URLs
  Those URLs will contain your bot-token. This may not be what you want.
  For now there is no secure way to relay GIF/stickers/documents without seeing your token.

## Bugfix
* irc: detect charset and try to convert it to utf-8 before sending it to other bridges. #209 #210
* slack: Remove label from URLs (slack). #205
* slack: Relay <>& correctly to other bridges #215
* steam: Fix channel id bug in steam (channels are off by 0x18000000000000)
* general: various improvements
* general: samechannelgateway now relays messages correct again #207


# v0.16.0-rc2
## Breaking Changes
* URL,UseAPI,BindAddress is deprecated. Your config has to be updated.
  * URL => WebhookURL
  * BindAddress => WebhookBindAddress
  * UseAPI => removed 
  This change allows you to specify a WebhookURL and a token (slack,discord), so that
  messages will be sent with the webhook, but received via the token (API)
  If you have not specified WebhookURL and WebhookBindAddress the API (login or token) 
  will be used automatically. (no need for UseAPI)

## Bugfix since rc1
* steam: Fix channel id bug in steam (channels are off by 0x18000000000000)
* telegram: Add UseInsecureURL option for telegram (default false)
  WARNING! If enabled this will relay GIF/stickers/documents and other attachments as URLs
  Those URLs will contain your bot-token. This may not be what you want.
  For now there is no secure way to relay GIF/stickers/documents without seeing your token.
* irc: detect charset and try to convert it to utf-8 before sending it to other bridges. #209 #210
* general: various improvements


# v0.16.0-rc1
## Breaking Changes
* URL,UseAPI,BindAddress is deprecated. Your config has to be updated.
  * URL => WebhookURL
  * BindAddress => WebhookBindAddress
  * UseAPI => removed 
  This change allows you to specify a WebhookURL and a token (slack,discord), so that
  messages will be sent with the webhook, but received via the token (API)
  If you have not specified WebhookURL and WebhookBindAddress the API (login or token) 
  will be used automatically. (no need for UseAPI)

## New features
* steam: New protocol support added (http://store.steampowered.com/)
* discord: WebhookURL posting support added (thanks @saury07) #204
  Discord API does not allow to change the name of the user posting, but webhooks does.

## Bugfix
* general: samechannelgateway now relays messages correct again #207
* slack: Remove label from URLs (slack). #205

# v0.15.0
## New features
* general: add option IgnoreMessages for all protocols (see mattebridge.toml.sample)
  Messages matching these regexp will be ignored and not sent to other bridges
  e.g. IgnoreMessages="^~~ badword"
* telegram: add support for sticker/video/photo/document #184

## Changes
* api: add userid to each message #200

## Bugfix
* discord: fix crash in memberupdate #198
* mattermost: Fix incorrect behaviour of EditDisable (mattermost). Fixes #197 
* irc: Do not relay join/part of ourselves (irc). Closes #190 
* irc: make reconnections more robust. #153
* gitter: update library, fixes possible crash

# v0.14.0
## New features
* api: add token authentication
* mattermost: add support for mattermost 3.10.0

## Changes
* api: gateway name is added in JSON messages
* api: lowercase JSON keys
* api: channel name isn't needed in config #195

## Bugfix
* discord: Add hashtag to channelname (when translating from id) (discord)
* mattermost: Fix a panic. #186
* mattermost: use teamid cache if possible. Fixes a panic
* api: post valid json. #185
* api: allow reuse of api in different gateways. #189
* general: Fix utf-8 issues for {NOPINGNICK}. #193

# v0.13.0
## New features
* irc: Limit message length. ```MessageLength=400```
  Maximum length of message sent to irc server. If it exceeds <message clipped> will be add to the message.
* irc: Add NOPINGNICK option. 
  The string "{NOPINGNICK}" (case sensitive) will be replaced by the actual nick / username, but with a ZWSP inside the nick, so the irc user with the same nick won't get pinged.   
  See https://github.com/42wim/matterbridge/issues/175 for more information

## Bugfix
* slack: Fix sending to different channels on same account (slack). Closes #177
* telegram: Fix incorrect usernames being sent. Closes #181


# v0.12.1
## New features
* telegram: Add UseFirstName option (telegram). Closes #144
* matrix: Add NoHomeServerSuffix. Option to disable homeserver on username (matrix). Closes #160.

## Bugfix
* xmpp: Add Compatibility for Cisco Jabber (xmpp) (#166)
* irc: Fix JoinChannel argument to use IRC channel key (#172)
* discord: Fix possible crash on nil (discord)
* discord: Replace long ids in channel metions (discord). Fixes #174

# v0.12.0
## Changes
* general: edited messages are now being sent by default on discord/mattermost/telegram/slack. See "New Features"

## New features
* general: add support for edited messages. 
  Add new keyword EditDisable (false/true), default false. Which means by default edited messages will be sent to other bridges.
  Add new keyword EditSuffix , default "". You can change this eg to "(edited)", this will be appended to every edit message.
* mattermost: support mattermost v3.9.x
* general: Add support for HTTP{S}_PROXY env variables (#162)
* discord: Strip custom emoji metadata (discord). Closes #148

## Bugfix
* slack: Ignore error on private channel join (slack) Fixes #150 
* mattermost: fix crash on reconnects when server is down. Closes #163
* irc: Relay messages starting with ! (irc). Closes #164

# v0.11.0
## New features
* general: reusing the same account on multiple gateways now also reuses the connection.
  This is particuarly useful for irc. See #87
* general: the Name is now REQUIRED and needs to be UNIQUE for each gateway configuration
* telegram:  Support edited messages (telegram). See #141
* mattermost: Add support for showing/hiding join/leave messages from mattermost. Closes #147
* mattermost: Reconnect on session removal/timeout (mattermost)
* mattermost: Support mattermost v3.8.x
* irc:  Rejoin channel when kicked (irc).

## Bugfix
* mattermost: Remove space after nick (mattermost). Closes #142
* mattermost: Modify iconurl correctly (mattermost).
* irc: Fix join/leave regression (irc)

# v0.10.3
## Bugfix
* slack: Allow bot tokens for now without warning (slack). Closes #140 (fixes user_is_bot message on channel join)

# v0.10.2
## New features
* general: gops agent added. Allows for more debugging. See #134
* general: toml inline table support added for config file

## Bugfix
* all: vendored libs updated

## Changes
* general: add more informative messages on startup

# v0.10.1
## Bugfix
* gitter: Fix sending messages on new channel join.

# v0.10.0
## New features
* matrix: New protocol support added (https://matrix.org)
* mattermost: works with mattermost release v3.7.0
* discord: Replace role ids in mentions to role names (discord). Closes #133

## Bugfix
* mattermost: Add ReadTimeout to close lingering connections (mattermost). See #125
* gitter: Join rooms not already joined by the bot (gitter). See #135
* general: Fail when bridge is unable to join a channel (general)

## Changes
* telegram: Do not use HTML parsemode by default. Set ```MessageFormat="HTML"``` to use it. Closes #126

# v0.9.3
## New features
* API: rest interface to read / post messages (see API section in matterbridge.toml.sample)

## Bugfix
* slack: fix receiving messages from private channels #118
* slack: fix echo when using webhooks #119
* mattermost: reconnecting should work better now
* irc: keeps reconnecting (every 60 seconds) now after ping timeout/disconnects.

# v0.9.2
## New features
* slack: support private channels #118

## Bugfix
* general: make ignorenicks work again #115
* telegram: fix receiving from channels and groups #112
* telegram: use html for username
* telegram: use ```unknown``` as username when username is not visible.
* irc: update vendor (fixes some crashes) #117
* xmpp: fix tls by setting ServerName #114

# v0.9.1
## New features
* Rocket.Chat: New protocol support added (https://rocket.chat)
* irc: add channel key support #27 (see matterbrige.toml.sample for example)
* xmpp: add SkipTLSVerify #106

## Bugfix
* general: Exit when a bridge fails to start
* mattermost: Check errors only on first connect. Keep retrying after first connection succeeds. #95
* telegram: fix missing username #102
* slack: do not use API functions in webhook (slack) #110

# v0.9.0
## New features
* Telegram: New protocol support added (https://telegram.org)
* Hipchat: Add sample config to connect to hipchat via xmpp
* discord: add "Bot " tag to discord tokens automatically
* slack: Add support for dynamic Iconurl #43
* general: Add ```gateway.inout``` config option for bidirectional bridges #85
* general: Add ```[general]``` section so that ```RemoteNickFormat``` can be set globally

## Bugfix
* general: when using samechannelgateway NickFormat get doubled by the NICK #77
* general: fix ShowJoinPart for messages from irc bridge #72
* gitter: fix high cpu usage #89
* irc: fix !users command #78
* xmpp: fix keepalive
* xmpp: do not relay delayed/empty messages
* slack: Replace id-mentions to usernames #86 
* mattermost: fix public links not working (API changes)

# v0.8.1
## Bugfix
* general: when using samechannelgateway NickFormat get doubled by the NICK #77
* irc: fix !users command #78

# v0.8.0
Release because of breaking mattermost API changes
## New features
* Supports mattermost v3.5.0

# v0.7.1
## Bugfix
* general: when using samechannelgateway NickFormat get doubled by the NICK #77
* irc: fix !users command #78

# v0.7.0
## Breaking config changes from 0.6 to 0.7
Matterbridge now uses TOML configuration (https://github.com/toml-lang/toml)
See matterbridge.toml.sample for an example

## New features
### General
* Allow for bridging the same type of bridge, which means you can eg bridge between multiple mattermosts.
* The bridge is now actually a gateway which has support multiple in and out bridges. (and supports multiple gateways).
* Discord support added. See matterbridge.toml.sample for more information.
* Samechannelgateway support added, easier configuration for 1:1 mapping of protocols with same channel names. #35
* Support for override from environment variables. #50
* Better debugging output.
* discord: New protocol support added. (http://www.discordapp.com)
* mattermost: Support attachments.
* irc: Strip colors. #33
* irc: Anti-flooding support. #40
* irc: Forward channel notices.

## Bugfix
* irc: Split newlines. #37
* irc: Only respond to nick related notices from nickserv.
* irc: Ignore queries send to the bot.
* irc: Ignore messages from ourself.
* irc: Only output the "users on irc information" when asked with "!users".
* irc: Actually wait until connection is complete before saying it is.
* mattermost: Fix mattermost channel joins.
* mattermost: Drop messages not from our team.
* slack: Do not panic on non-existing channels.
* general: Exit when a bridge fails to start.

# v0.6.1
## New features
* Slack support added.  See matterbridge.conf.sample for more information

## Bugfix
* Fix 100% CPU bug on incorrect closed connections

# v0.6.0-beta2
## New features
* Gitter support added.  See matterbridge.conf.sample for more information

# v0.6.0-beta1
## Breaking changes from 0.5 to 0.6
### commandline
* -plus switch deprecated. Use ```Plus=true``` or ```Plus``` in ```[general]``` section

### IRC section
* ```Enabled``` added (default false)  
Add ```Enabled=true``` or ```Enabled``` to the ```[IRC]``` section if you want to enable the IRC bridge

### Mattermost section
* ```Enabled``` added (default false)  
Add ```Enabled=true``` or ```Enabled``` to the ```[mattermost]``` section if you want to enable the mattermost bridge

### General section
* Use ```Plus=true``` or ```Plus``` in ```[general]``` section to enable the API version of matterbridge

## New features
* Matterbridge now bridges between any specified protocol (not only mattermost anymore) 
* XMPP support added.  See matterbridge.conf.sample for more information
* RemoteNickFormat {BRIDGE} variable added  
You can now add the originating bridge to ```RemoteNickFormat```  
eg ```RemoteNickFormat="[{BRIDGE}] <{NICK}> "```


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
Removed, can be done by using ```RemoteNickFormat="<{NICK}> "```

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
