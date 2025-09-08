# Settings

On this page you will find information about global matterbridge settings. Most of these settings can be applied to specific gateways, and those marked **GENERAL** can be placed in the `[general]` configuration dictionary for setting it across all gateways.

> [!TIP]
> Settings specific to a certain protocol can be found in the corresponding folder in the
> [protocols/](protocols/) documentation folder. API settings for can be found in the
> [api/](api/) documentation folder.

# Info

* **OPTIONAL:** this setting isn't enabled by default.
* **RELOADABLE:** this setting can be reloaded by editing the configuration. No restart of matterbridge is required.
* **ALL:** this setting is usable with all bridges.
* **GENERAL:** this setting can be also set under `[general]` which means it's active for all bridges.

# Shared
Only settings which have the `ALL` setting are usable for all bridges.

## EditDisable
Disable sending of edits to other bridges

Setting: OPTIONAL, RELOADABLE \
Format: boolean \
Example: enable it

`EditDisable=true`

## EditSuffix
Message to be appended to every edited message

Setting: OPTIONAL, RELOADABLE \
Format: string \
Example: 

`EditSuffix=" (edited)"`

## IgnoreMessages
Messages you want to ignore.\
Messages matching these regexp will be ignored and not sent to other bridges.\
See https://regex-golang.appspot.com/assets/html/index.html for more regex info.

Setting: OPTIONAL, RELOADABLE, ALL \
Format: space seperated regex string \
Example: ignores messages starting with ~~ or messages containing badword

`IgnoreMessages="^~~ badword"`

## IgnoreNicks
Nicks you want to ignore.\
Messages from those users will not be sent to other bridges.

Setting: OPTIONAL, RELOADABLE, ALL \
Format: space seperated regex string \
Example: ignore messages from ircspammer1 and ircspammer2

`IgnoreNicks="ircspammer1 ircspammer2"`

## Label
Extra label that can be used in the `RemoteNickFormat`

Setting: OPTIONAL, RELOADABLE, ALL \
Format: string \
Example: add the label `mychat`

`Label="mychat"`

## PrefixMessagesWithNick
Whether to prefix messages from other bridges to with the sender's nick.
Useful if username overrides for incoming webhooks isn't enabled.
 If you set PrefixMessagesWithNick to true, each message
from other bridges will by default be prefixed by "bridge-" + nick. You can,
however, modify how the messages appear, by setting (and modifying) `RemoteNickFormat`

Setting: OPTIONAL, RELOADABLE \
Format: boolean \
Example: 

`PrefixMessagesWithNick=true`


## RemoteNickFormat 
Defines how remote users appear on this bridge. \
The string "{NICK}" (case sensitive) will be replaced by the actual nick / username. \
The string "{BRIDGE}" (case sensitive) will be replaced by the sending bridge. \
The string "{LABEL}" (case sensitive) will be replaced by label= field of the sending bridge. \
The string "{PROTOCOL}" (case sensitive) will be replaced by the protocol used by the bridge. \
The string "{GATEWAY}" (case sensitive) will be replaced by the origin gateway name that is replicating the message. \
The string "{CHANNEL}" (case sensitive) will be replaced by the origin channel name used by the bridge. \
The string "{TENGO}" (case sensitive) will be replaced by the output of the RemoteNickFormat script under `[tengo]` \
The string "{NOPINGNICK}" (case sensitive) will be replaced by the actual nick / username, but with a ZWSP inside the nick, so the irc user with the same nick won't get pinged. See https://github.com/42wim/matterbridge/issues/175 for more information

Setting: OPTIONAL, RELOADABLE, GENERAL, ALL \
Format: string \
Example: add PROTOCOL and NICK

`RemoteNickFormat="[{PROTOCOL}] <{NICK}> "`

## ReplaceMessages
Messages you want to replace. \
It replaces outgoing messages from the bridge. \
So you need to place it by the sending bridge definition.\
Regular expressions supported.

Setting: OPTIONAL, RELOADABLE, ALL \
Format: [ ["from1","to1"],["from2","to2"] ] \
Example: this replaces cat => dog and sleep => awake

`ReplaceMessages=[ ["cat","dog"], ["sleep","awake"] ]`

## ReplaceNicks
Nicks you want to replace. \
See `ReplaceMessages` for syntax

Setting: OPTIONAL, RELOADABLE, ALL \
Format: [ ["from1","to1"],["from2","to2"] ] \
Example: this replaces user-- => user

`ReplaceNicks=[ ["user--","user"] ]`

## ShowJoinPart
Enable to show users joins/parts from other bridges \
Currently works for messages from the following bridges: irc, mattermost, slack

Setting: OPTIONAL, RELOADABLE, ALL \
Format: boolean \
Example: enable it

`ShowJoinPart=true`

## ShowTopicChange
Enable to show topic changes from other bridges. \
Only works hiding/show topic changes from slack bridge for now. 

Setting: OPTIONAL, RELOADABLE, ALL \
Format: boolean \
Example: enable it

`ShowTopicChange=true`

## SkipTLSVerify
Enable to not verify the certificate on your server.
e.g. when using selfsigned certificates

Setting: OPTIONAL \
Format: boolean \
Example: enable it

`SkipTLSVerify=true`

## StripNick
StripNick only allows alphanumerical nicks. See https://github.com/42wim/matterbridge/issues/285
It will strip other characters from the nick

Setting: OPTIONAL, RELOADABLE, GENERAL, ALL \
Format: boolean \
Example: enable it

`StripNick=true`

## UseLocalAvatar

UseLocalAvatar specifies source bridges for which an avatar should be 'guessed' when an incoming message has no avatar. This works by comparing the username of the message to an existing Discord user, and using the avatar of the Discord user. (Substitute "Discord" with another platform, if used on another platform.)

On Discord, this only works if `WebhookURL` is set (AND the message has no avatar).

At the moment, this setting is only available for Discord.

Setting: OPTIONAL, RELOADABLE  
Format: `["gateway.account1", "gateway.account2", "gateway"]`  
Example: `["irc"]` - this example will guess the avatar coming from the `irc` platform  

# General

Configuration that can be set under `[general]`

## IgnoreFailureOnStart 
Allows you to ignore failing bridges on startup. 
Matterbridge will disable the failed bridge and continue with the other ones. \
Context: https://github.com/42wim/matterbridge/issues/455

Setting: OPTIONAL, RELOADABLE, GENERAL \
Format: boolean \
Example: enable it

`IgnoreFailureOnStart=true`

## LogFile

LogFile defines the location of a file to write logs into, rather than stdout.
Logging will still happen on stdout if the file cannot be open for #writing, or if the value is empty. 
Note that the log won't roll, so you might want to use logrotate(8) with this feature.

Setting: OPTIONAL, GENERAL \
Format: string \
Example:

`LogFile="/var/log/matterbridge.log"`

## MediaDownloadBlacklist 
Allows you to blacklist specific files from being downloaded.
Filenames matching these regexp will not be download/uploaded to the mediaserver. \
You can use regex for this, see https://regex-golang.appspot.com/assets/html/index.html for more regex info

Setting: OPTIONAL, RELOADABLE, GENERAL \
Format: string array \
Example: do not upload html and htm extension

`MediaDownloadBlacklist=[".html$",".htm$"]`

## MediaDownloadPath
MediaDownloadPath is the filesystem path where the media file will be placed, instead of uploaded, for if Matterbridge has write access to the directory your webserver is serving. \
It is an alternative to MediaServerUpload.
More information https://github.com/42wim/matterbridge/wiki/Mediaserver-setup-%28advanced%29

Setting: OPTIONAL, RELOADABLE, GENERAL \
Format: string \
Example: 

`MediaDownloadPath="/srv/http/yourserver.com/public/download"`

## MediaDownloadSize
Maximum size in bytes matterbridge will download for use with (`MediaServerUpload` and `MediaDownloadPath`)

Setting: OPTIONAL, RELOADABLE, GENERAL \
Format: int \
Default: 10000000 (1 megabyte) \
Example:

`MediaDownloadSize=1000000`


## MediaServerDownload
The MediaServerDownload will be used so that bridges without native uploading support:
gitter, irc and xmpp will be shown links to the files on MediaServerDownload
More information https://github.com/42wim/matterbridge/wiki/Mediaserver-setup-%28advanced%29

Setting: OPTIONAL, RELOADABLE, GENERAL \
Format: string \
Example: 

`MediaServerDownload="https://youserver.com/download"`


## MediaServerUpload 
Used for uploading images/files/video to a remote "mediaserver" (a webserver like caddy for example). \
When configured images/files uploaded on bridges like mattermost, slack, telegram will be
downloaded and uploaded again to MediaServerUpload URL.
More information https://github.com/42wim/matterbridge/wiki/Mediaserver-setup-%28advanced%29

Setting: OPTIONAL, RELOADABLE, GENERAL \
Format: string \
Example: 

`MediaServerUpload="https://user:pass@yourserver.com/upload"`

## RemoteNickFormat 
See [RemoteNickFormat](#RemoteNickFormat)
