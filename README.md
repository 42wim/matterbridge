# matterbridge

Simple bridge between mattermost and IRC. Uses the in/outgoing webhooks.  
Relays public channel messages between mattermost and IRC.  

Work in progress. 

## building
Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH] (https://golang.org/doc/code.html#GOPATH)

```
cd $GOPATH
go get https://github.com/42wim/matterbridge
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

Matterbridge will:
* start a webserver listening on the port specified in the configuration.
* connect to specified irc server and channel.
* send messages from mattermost to irc and vice versa, messages in mattermost will appear with irc-nick

## config
### matterbridge
matterbridge looks for matterbridge.conf in current directory.

Look at matterbridge.conf.sample for an example


```
[IRC]
server="irc.freenode.net"
port=6667
UseTLS=false
SkipTLSVerify=true
nick="matterbot"
channel="#matterbridge"

[mattermost]
#url is your incoming webhook url (account settings - integrations - incoming webhooks)
url="http://mattermost.yourdomain.com/hooks/incomingwebhookkey"  
#port the bridge webserver will listen on
port=9999
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

