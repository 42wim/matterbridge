# Breaking changes from 0.4 to 0.5 for matterbridge (webhooks version)
## IRC section
### Server
Port removed, added to server
```
server="irc.freenode.net"
port=6667
```
changed to
```
server="irc.freenode.net:6667"
```
### Channel
Removed see Channels section below

### UseSlackCircumfix=true
Removed, can be done by using ```RemoteNickFormat="<{NICK}> "```

## Mattermost section
### BindAddress
Port removed, added to BindAddress

```
BindAddress="0.0.0.0"
port=9999
```

changed to

```
BindAddress="0.0.0.0:9999"
```

### Token
Removed

## Channels section
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
