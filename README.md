# matterbridge

Simple bridge between mattermost and IRC. 

* Relays public channel messages between mattermost and IRC.
* Supports multiple mattermost and irc channels.
* Matterbridge -plus also works with private groups on your mattermost.

## Requirements:
* [Mattermost] (https://github.com/mattermost/platform/) 3.x (stable, not a dev build)
### Webhooks version
* Configured incoming/outgoing [webhooks](https://www.mattermost.org/webhooks/) on your mattermost instance.
### Plus (API) version
* A dedicated user(bot) on your mattermost instance.

## binaries
Binaries can be found [here] (https://github.com/42wim/matterbridge/releases/tag/v0.5)

## building
Go 1.6+ is required. Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH] (https://golang.org/doc/code.html#GOPATH)

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
Usage of ./matterbridge:
  -conf string
        config file (default "matterbridge.conf")
  -debug
        enable debug
  -plus
        running using API instead of webhooks
  -version
        show version
```

## config
### matterbridge
matterbridge looks for matterbridge.conf in current directory. (use -conf to specify another file)

Look at [matterbridge.conf.sample] (https://github.com/42wim/matterbridge/blob/master/matterbridge.conf.sample) for an example.

### mattermost
#### webhooks version
You'll have to configure the incoming and outgoing webhooks. 

* incoming webhooks
Go to "account settings" - integrations - "incoming webhooks".  
Choose a channel at "Add a new incoming webhook", this will create a webhook URL right below.  
This URL should be set in the matterbridge.conf in the [mattermost] section (see above)  

* outgoing webhooks
Go to "account settings" - integrations - "outgoing webhooks".  
Choose a channel (the same as the one from incoming webhooks) and fill in the address and port of the server matterbridge will run on.  

e.g. http://192.168.1.1:9999 (192.168.1.1:9999 is the BindAddress specified in [mattermost] section of matterbridge.conf)

#### plus version
You'll have to create a new dedicated user on your mattermost instance.
Specify the login and password in [mattermost] section of matterbridge.conf
