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
