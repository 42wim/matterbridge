# matterbridge

Simple bridge between mattermost and IRC. 

* Relays public channel messages between mattermost and IRC.
* Supports multiple mattermost and irc channels.
* Matterbridge -plus also works with private groups on your mattermost.

This project has now [matterbridge-plus](https://github.com/42wim/matterbridge-plus/) merged in.  
Breaking changes for matterbridge can be found in [migration](https://github.com/42wim/matterbridge/blob/master/migration.md)  
Look at [matterbridge.conf.sample] (https://github.com/42wim/matterbridge/blob/master/matterbridge.conf.sample) for an example.

Configuration changes since v0.5.0 can be found in [changelog.md] (https://github.com/42wim/matterbridge/blob/master/changelog.md)

## Requirements:
* [Mattermost] (https://github.com/mattermost/platform/)

### Compatibility
* Matterbridge v0.6.0 works with mattermost 3.3.0 and higher [3.3.0 release](https://github.com/mattermost/platform/releases/tag/v3.3.0)
* Matterbridge v0.5.0 works with mattermost 3.0.0 - 3.2.0 [3.2.0 release](https://github.com/mattermost/platform/releases/tag/v3.2.0)


### Webhooks version
* Configured incoming/outgoing [webhooks](https://www.mattermost.org/webhooks/) on your mattermost instance.

### Plus (API) version
* A dedicated user(bot) on your mattermost instance.

## binaries
Binaries can be found [here] (https://github.com/42wim/matterbridge/releases/)
* For use with mattermost 3.3.0 [v0.6.0-beta1](https://github.com/42wim/matterircd/releases/tag/v0.6.0-beta1)
* For use with mattermost 3.0.0-3.2.0 [v0.5.0](https://github.com/42wim/matterircd/releases/tag/v0.5.0)

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
        running using API instead of webhooks (deprecated, set Plus flag in [general] config)
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

## FAQ
Please look at [matterbridge.conf.sample] (https://github.com/42wim/matterbridge/blob/master/matterbridge.conf.sample) for more information first. 
### Mattermost doesn't show the IRC nicks
If you're running the webhooks version, this can be fixed by either:
* enabling "override usernames". See [mattermost documentation](http://docs.mattermost.com/developer/webhooks-incoming.html#enabling-incoming-webhooks)
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.conf.

If you're running the plus version you'll need to:
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.conf.

Also look at the ```RemoteNickFormat``` setting.
