# matterbridge
Click on one of the badges below to join the chat   

[![Gitter](https://img.shields.io/gitter/room/nwjs/nw.js.svg)](https://gitter.im/42wim/matterbridge) [![Join the IRC chat at https://webchat.freenode.net/?channels=matterbridgechat](https://img.shields.io/badge/IRC-matterbridgechat-green.svg)](https://webchat.freenode.net/?channels=matterbridgechat) [![Discord](https://img.shields.io/badge/discord-matterbridge-green.svg)](https://discord.gg/AkKPtrQ) [![Matrix](https://img.shields.io/badge/matrix-matterbridge-green.svg)](https://riot.im/app/#/room/#matterbridge:matrix.org) [![Slack](https://img.shields.io/badge/slack-matterbridgechat-green.svg)](https://join.slack.com/matterbridgechat/shared_invite/MjEwODMxNjU1NDMwLTE0OTk2MTU3NTMtMzZkZmRiNDZhOA) [![Mattermost](https://img.shields.io/badge/mattermost-matterbridge-green.svg)](https://framateam.org/signup_user_complete/?id=tfqm33ggop8x3qgu4boeieta6e)

![matterbridge.gif](https://s15.postimg.org/qpjhp6y3f/matterbridge.gif)

Simple bridge between Mattermost, IRC, XMPP, Gitter, Slack, Discord, Telegram, Rocket.Chat, Hipchat(via xmpp), Matrix and Steam.
Has a REST API.

# Table of Contents
 * [Features](#features)
 * [Requirements](#requirements)
 * [Installing](#installing)
   * [Binaries](#binaries)
   * [Building](#building)
 * [Configuration](#configuration)
   * [Howto](https://github.com/42wim/matterbridge/wiki/How-to-create-your-config)
   * [Examples](#examples) 
 * [Running](#running)
   * [Docker](#docker)
 * [Changelog](#changelog)
 * [FAQ](#faq)
 * [Thanks](#thanks)

# Features
* Relays public channel messages between multiple mattermost, IRC, XMPP, Gitter, Slack, Discord, Telegram, Rocket.Chat, Hipchat (via xmpp), Matrix and Steam. 
  Pick and mix.
* Matterbridge can also work with private groups on your mattermost/slack.
* Allow for bridging the same bridges, which means you can eg bridge between multiple mattermosts.
* The bridge is now a gateway which has support multiple in and out bridges. (and supports multiple gateways).
* REST API to read/post messages to bridges (WIP).

# Requirements
Accounts to one of the supported bridges
* [Mattermost](https://github.com/mattermost/platform/) 3.5.x - 3.10.x, 4.0.x
* [IRC](http://www.mirc.com/servers.html)
* [XMPP](https://jabber.org)
* [Gitter](https://gitter.im)
* [Slack](https://slack.com)
* [Discord](https://discordapp.com)
* [Telegram](https://telegram.org)
* [Hipchat](https://www.hipchat.com)
* [Rocket.chat](https://rocket.chat)
* [Matrix](https://matrix.org)
* [Steam](https://store.steampowered.com/)

# Installing
## Binaries
* Latest stable release [v0.16.1](https://github.com/42wim/matterbridge/releases/tag/v0.16.1)
* Development releases (follows master) can be downloaded [here](https://dl.bintray.com/42wim/nightly/)

## Building
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

# Configuration
## Basic configuration
See [howto](https://github.com/42wim/matterbridge/wiki/How-to-create-your-config) for a step by step walkthrough for creating your configuration.

## Advanced configuration
* [matterbridge.toml.sample](https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample) for documentation and an example.

## Examples 
### Bridge mattermost (off-topic) - irc (#testing)
```
[irc]
    [irc.freenode]
    Server="irc.freenode.net:6667"
    Nick="yourbotname"

[mattermost]
    [mattermost.work]
    useAPI=true
    Server="yourmattermostserver.tld"
    Team="yourteam"
    Login="yourlogin"
    Password="yourpass"
    PrefixMessagesWithNick=true

[[gateway]]
name="mygateway"
enable=true
    [[gateway.inout]]
    account="irc.freenode"
    channel="#testing"

    [[gateway.inout]]
    account="mattermost.work"
    channel="off-topic"
```

### Bridge slack (#general) - discord (general)
```
[slack]
[slack.test]
useAPI=true
Token="yourslacktoken"
PrefixMessagesWithNick=true

[discord]
[discord.test]
Token="yourdiscordtoken"
Server="yourdiscordservername"

[general]
RemoteNickFormat="[{PROTOCOL}/{BRIDGE}] <{NICK}> "

[[gateway]]
    name = "mygateway"
    enable=true

    [[gateway.inout]]
    account = "discord.test"
    channel="general"

    [[gateway.inout]]
    account ="slack.test"
    channel = "general"
```

# Running
1) Copy the matterbridge.toml.sample to matterbridge.toml 
2) Edit matterbridge.toml with the settings for your environment. 
3) Now you can run matterbridge.  (```./matterbridge```)   

(Matterbridge will only look for the config file in your current directory, if it isn't there specify -conf "/path/toyour/matterbridge.toml")

```
Usage of ./matterbridge:
  -conf string
        config file (default "matterbridge.toml")
  -debug
        enable debug
  -gops
        enable gops agent
  -version
        show version
```

## Docker
Create your matterbridge.toml file locally eg in ```/tmp/matterbridge.toml```
```
docker run -ti -v /tmp/matterbridge.toml:/matterbridge.toml 42wim/matterbridge
```

# Changelog
See [changelog.md](https://github.com/42wim/matterbridge/blob/master/changelog.md)

# FAQ

Please look at [matterbridge.toml.sample](https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample) for more information first.

## Mattermost doesn't show the IRC nicks
If you're running the webhooks version, this can be fixed by either:
* enabling "override usernames". See [mattermost documentation](http://docs.mattermost.com/developer/webhooks-incoming.html#enabling-incoming-webhooks)
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.toml.

If you're running the API version you'll need to:
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.toml.

Also look at the ```RemoteNickFormat``` setting.


# Thanks
Matterbridge wouldn't exist without these libraries:
* discord - https://github.com/bwmarrin/discordgo
* echo - https://github.com/labstack/echo
* gitter - https://github.com/sromku/go-gitter
* gops - https://github.com/google/gops
* irc - https://github.com/thoj/go-ircevent
* mattermost - https://github.com/mattermost/platform
* matrix - https://github.com/matrix-org/gomatrix
* slack - https://github.com/nlopes/slack
* steam - https://github.com/Philipp15b/go-steam
* telegram - https://github.com/go-telegram-bot-api/telegram-bot-api
* xmpp - https://github.com/mattn/go-xmpp
