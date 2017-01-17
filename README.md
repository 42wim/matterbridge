# matterbridge
![matterbridge.gif](https://s15.postimg.org/qpjhp6y3f/matterbridge.gif)

Simple bridge between mattermost, IRC, XMPP, Gitter, Slack, Discord, Telegram, Rocket.Chat and Hipchat(via xmpp).

* Relays public channel messages between multiple mattermost, IRC, XMPP, Gitter, Slack, Discord, Telegram, Rocket.Chat and Hipchat (via xmpp). Pick and mix.
* Supports multiple channels.
* Matterbridge can also work with private groups on your mattermost.
* Allow for bridging the same bridges, which means you can eg bridge between multiple mattermosts.
* The bridge is now a gateway which has support multiple in and out bridges. (and supports multiple gateways).

Look at [matterbridge.toml.sample] (https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample) for documentation and an example.
Look at [matterbridge.toml.simple] (https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.simple) for a simple example.


## Changelog
Since v0.7.0 the configuration has changed. More details in [changelog.md] (https://github.com/42wim/matterbridge/blob/master/changelog.md)

## Requirements
Accounts to one of the supported bridges
* [Mattermost] (https://github.com/mattermost/platform/)
* [IRC] (http://www.mirc.com/servers.html)
* [XMPP] (https://jabber.org)
* [Gitter] (https://gitter.im)
* [Slack] (https://slack.com)
* [Discord] (https://discordapp.com)
* [Telegram] (https://telegram.org)
* [Hipchat] (https://www.hipchat.com)
* [Rocket.chat] (https://rocket.chat)

## Docker
Create your matterbridge.toml file locally eg in ```/tmp/matterbridge.toml```
```
docker run -ti -v /tmp/matterbridge.toml:/matterbridge.toml 42wim/matterbridge
```

## binaries
Binaries can be found [here] (https://github.com/42wim/matterbridge/releases/)
* For use with mattermost 3.5.x - 3.6.0 [v0.9.1](https://github.com/42wim/matterircd/releases/tag/v0.9.1)
* For use with mattermost 3.3.0 - 3.4.0 [v0.7.1](https://github.com/42wim/matterircd/releases/tag/v0.7.1)

## Compatibility
### Mattermost 
* Matterbridge v0.9.1 works with mattermost 3.5.x - 3.6.0 [3.6.0 release](https://github.com/mattermost/platform/releases/tag/v3.6.0)
* Matterbridge v0.7.1 works with mattermost 3.3.0 - 3.4.0 [3.4.0 release](https://github.com/mattermost/platform/releases/tag/v3.4.0)

#### Webhooks version
* Configured incoming/outgoing [webhooks](https://www.mattermost.org/webhooks/) on your mattermost instance.

#### API version
* A dedicated user(bot) on your mattermost instance.


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
        config file (default "matterbridge.toml")
  -debug
        enable debug
  -version
        show version
```

## config
### matterbridge
matterbridge looks for matterbridge.toml in current directory. (use -conf to specify another file)

Look at [matterbridge.toml.sample] (https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample) for an example.

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

## FAQ
Please look at [matterbridge.toml.sample] (https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample) for more information first. 
### Mattermost doesn't show the IRC nicks
If you're running the webhooks version, this can be fixed by either:
* enabling "override usernames". See [mattermost documentation](http://docs.mattermost.com/developer/webhooks-incoming.html#enabling-incoming-webhooks)
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.toml.

If you're running the plus version you'll need to:
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.toml.

Also look at the ```RemoteNickFormat``` setting.
