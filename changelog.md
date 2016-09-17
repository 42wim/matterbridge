# v0.6.0
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
