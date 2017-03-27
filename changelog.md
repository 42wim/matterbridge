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
