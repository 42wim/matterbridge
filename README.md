# matterbridge

Simple bridge between mattermost and IRC. Uses the in/outgoing webhooks.  
Relays public channel messages between mattermost and IRC.  

Requires mattermost 1.2.0+

There is also [matterbridge-plus] (https://github.com/42wim/matterbridge-plus) which uses the mattermost API and needs a dedicated user (bot). But requires no incoming/outgoing webhook setup. 

## binaries
Binaries can be found [here] (https://github.com/42wim/matterbridge/releases/tag/v0.3)

## building
Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH] (https://golang.org/doc/code.html#GOPATH)

```
cd $GOPATH
go get github.com/42wim/matterbridge
```

You should now have matterbridge binary in the bin directory:

```
$ ls bin/
matterbridge
```

## running
1) Copy the matterbridge.conf.sample to matterbridge.conf in the same directory as the matterbridge binary.  
2) Edit matterbridge.conf with the settings for your environment. See below for more config information.  
3) Now you can run matterbridge. 

```
Usage of matterbridge:
  -conf="matterbridge.conf": config file
```

Matterbridge will:
* start a webserver listening on the port specified in the configuration.
* connect to specified irc server and channel.
* send messages from mattermost to irc and vice versa, messages in mattermost will appear with irc-nick

## config
### matterbridge
matterbridge looks for matterbridge.conf in current directory. (use -conf to specify another file)

Look at matterbridge.conf.sample for an example


```
[IRC]
server="irc.freenode.net"
port=6667
UseTLS=false
SkipTLSVerify=true
nick="matterbot"
channel="#matterbridge"
UseSlackCircumfix=false

[mattermost]
#url is your incoming webhook url (account settings - integrations - incoming webhooks)
url="http://mattermost.yourdomain.com/hooks/incomingwebhookkey"  
#port the bridge webserver will listen on
port=9999
#address the webserver will bind to
BindAddress="0.0.0.0"
showjoinpart=true #show irc users joining and parting
#the token you get from the outgoing webhook in mattermost. If empty no token check will be done.
#if you use multiple IRC channel (see below, this must be empty!)
token=yourtokenfrommattermost
#disable certificate checking (selfsigned certificates)
#SkipTLSVerify=true
#whether to prefix messages from IRC to mattermost with the sender's nick. Useful if username overrides for incoming webhooks isn't enabled on the mattermost server
PrefixMessagesWithNick=false
#how to format the list of IRC nicks when displayed in mattermost. Possible options are "table" and "plain"
NickFormatter=plain
#how many nicks to list per row for formatters that support this
NicksPerRow=4
#Freenode nickserv
NickServNick="nickserv"
#Password for nickserv
NickServPassword="secret"

#multiple channel config
#token you can find in your outgoing webhook
[Token "outgoingwebhooktoken1"] 
IRCChannel="#off-topic"
MMChannel="off-topic"

[Token "outgoingwebhooktoken2"]
IRCChannel="#testing"
MMChannel="testing"

[general]
#request your API key on https://github.com/giphy/GiphyAPI. This is a public beta key
GiphyApiKey="dc6zaTOxFJmzC"
```

### mattermost
You'll have to configure the incoming en outgoing webhooks. 

* incoming webhooks
Go to "account settings" - integrations - "incoming webhooks".  
Choose a channel at "Add a new incoming webhook", this will create a webhook URL right below.  
This URL should be set in the matterbridge.conf in the [mattermost] section (see above)  

* outgoing webhooks
Go to "account settings" - integrations - "outgoing webhooks".  
Choose a channel (the same as the one from incoming webhooks) and fill in the address and port of the server matterbridge will run on.  

e.g. http://192.168.1.1:9999 (9999 is the port specified in [mattermost] section of matterbridge.conf)

