# matterbridge
Click on one of the badges below to join the chat   

[![Gitter](https://img.shields.io/gitter/room/nwjs/nw.js.svg?colorB=42f4242)](https://gitter.im/42wim/matterbridge) [![Join the IRC chat at https://webchat.freenode.net/?channels=matterbridgechat](https://img.shields.io/badge/IRC-matterbridgechat-green.svg?colorB=42f4242)](https://webchat.freenode.net/?channels=matterbridgechat) [![Discord](https://img.shields.io/badge/discord-matterbridge-green.svg?colorB=42f4242)](https://discord.gg/AkKPtrQ) [![Matrix](https://img.shields.io/badge/matrix-matterbridge-green.svg?colorB=42f4242)](https://riot.im/app/#/room/#matterbridge:matrix.org) [![Slack](https://img.shields.io/badge/slack-matterbridgechat-green.svg?colorB=42f4242)](https://join.slack.com/matterbridgechat/shared_invite/MjEwODMxNjU1NDMwLTE0OTk2MTU3NTMtMzZkZmRiNDZhOA) [![Mattermost](https://img.shields.io/badge/mattermost-matterbridge-green.svg?colorB=42f4242)](https://framateam.org/signup_user_complete/?id=tfqm33ggop8x3qgu4boeieta6e) [![Xmpp](https://img.shields.io/badge/xmpp-matterbridge@conference.jabber.de-green.svg?colorB=42f4242)](https://inverse.chat) [![Twitch](https://img.shields.io/badge/twitch-matterbridge-green.svg?colorB=42f4242)](https://www.twitch.tv/matterbridge) [![Zulip](https://img.shields.io/badge/zulip-matterbridge-green.svg?colorB=42f4242)](https://matterbridge.zulipchat.com/register/)

[![Download stable](https://img.shields.io/github/release/42wim/matterbridge.svg?label=download%20stable)](https://github.com/42wim/matterbridge/releases/latest) [![Download dev](https://img.shields.io/bintray/v/42wim/nightly/Matterbridge.svg?label=download%20dev&colorB=007ec6)](https://bintray.com/42wim/nightly/Matterbridge/_latestVersion)

![matterbridge.gif](https://github.com/42wim/matterbridge/blob/master/img/matterbridge.gif)

Simple bridge between IRC, XMPP, Gitter, Mattermost, Slack, Discord, Telegram, Rocket.Chat, Hipchat(via xmpp), Matrix, Steam, ssh-chat and Zulip
Has a REST API.   
Minecraft server chat support via [MatterLink](https://github.com/elytra/MatterLink)

**Mattermost isn't required to run matterbridge. It bridges between any supported protocol.**   
(The name matterbridge is a remnant when it was only bridging mattermost)

# Table of Contents
 * [Features](https://github.com/42wim/matterbridge/wiki/Features)
   * [API](#API)
 * [Requirements](#requirements)
 * [Screenshots](https://github.com/42wim/matterbridge/wiki/)
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
 * [Related projects](#related-projects)
 * [Articles](#articles)
 * [Thanks](#thanks)

# Features
* [Support bridging between any protocols](https://github.com/42wim/matterbridge/wiki/Features#support-bridging-between-any-protocols)
* [Support multiple gateways(bridges) for your protocols](https://github.com/42wim/matterbridge/wiki/Features#support-multiple-gatewaysbridges-for-your-protocols)
* [Message edits and deletes](https://github.com/42wim/matterbridge/wiki/Features#message-edits-and-deletes)
* Preserves threading when possible
* Remote configuration files
* [Attachment / files handling](https://github.com/42wim/matterbridge/wiki/Features#attachment--files-handling)
* [Username and avatar spoofing](https://github.com/42wim/matterbridge/wiki/Features#username-and-avatar-spoofing)
* [Private groups](https://github.com/42wim/matterbridge/wiki/Features#private-groups)
* [API](https://github.com/42wim/matterbridge/wiki/Features#api)

## API
The API is very basic at the moment and rather undocumented.

Used by at least 3 projects. Feel free to make a PR to add your project to this list.

* [MatterLink](https://github.com/elytra/MatterLink) (Matterbridge link for Minecraft Server chat)
* [pyCord](https://github.com/NikkyAI/pyCord) (crossplatform chatbot)
* [Mattereddit](https://github.com/bonehurtingjuice/mattereddit) (Reddit chat support)

# Requirements
Accounts to one of the supported bridges
* [Mattermost](https://github.com/mattermost/platform/) 3.8.x - 3.10.x, 4.x, 5.x
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
* [Twitch](https://twitch.tv)
* [Ssh-chat](https://github.com/shazow/ssh-chat)
* [Zulip](https://zulipchat.com)

# Screenshots
See https://github.com/42wim/matterbridge/wiki

# Installing
## Binaries
* Latest stable release [v1.11.3](https://github.com/42wim/matterbridge/releases/latest)
* Development releases (follows master) can be downloaded [here](https://dl.bintray.com/42wim/nightly/)  

## Building
Go 1.8+ is required. Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH](https://golang.org/doc/code.html#GOPATH).

After Go is setup, download matterbridge to your $GOPATH directory. 

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
    Server="yourmattermostserver.tld"
    Team="yourteam"
    Login="yourlogin"
    Password="yourpass"
    PrefixMessagesWithNick=true
    RemoteNickFormat="[{PROTOCOL}] <{NICK}> "

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

See [howto](https://github.com/42wim/matterbridge/wiki/How-to-create-your-config) for a step by step walkthrough for creating your configuration.

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

See [FAQ](https://github.com/42wim/matterbridge/wiki/FAQ)

Want to tip ? 
* eth: 0xb3f9b5387c66ad6be892bcb7bbc67862f3abc16f
* btc: 1N7cKHj5SfqBHBzDJ6kad4BzeqUBBS2zhs

# Related projects
* [matterbridge-heroku](https://github.com/cadecairos/matterbridge-heroku)
* [matterbridge config viewer](https://github.com/patcon/matterbridge-heroku-viewer)
* [matterbridge autoconfig](https://github.com/patcon/matterbridge-autoconfig)
* [matterlink](https://github.com/elytra/MatterLink)
* [mattereddit](https://github.com/bonehurtingjuice/mattereddit)
* [pyCord](https://github.com/NikkyAI/pyCord) (crossplatform chatbot)

# Articles
* https://mattermost.com/blog/connect-irc-to-mattermost/
* https://blog.valvin.fr/2016/09/17/mattermost-et-un-channel-irc-cest-possible/
* https://blog.brightscout.com/top-10-mattermost-integrations/
* http://bencey.co.nz/2018/09/17/bridge/
* https://www.algoo.fr/blog/2018/01/19/recouvrez-votre-liberte-en-quittant-slack-pour-un-mattermost-auto-heberge/
* https://kopano.com/blog/matterbridge-bridging-mattermost-chat/
* https://www.stitcher.com/s/?eid=52382713

# Thanks
[![Digitalocean](https://snag.gy/3LVifX.jpg)](https://www.digitalocean.com/) for sponsoring demo/testing droplets.

Matterbridge wouldn't exist without these libraries:
* discord - https://github.com/bwmarrin/discordgo
* echo - https://github.com/labstack/echo
* gitter - https://github.com/sromku/go-gitter
* gops - https://github.com/google/gops
* gozulipbot - https://github.com/ifo/gozulipbot
* irc - https://github.com/lrstanley/girc
* mattermost - https://github.com/mattermost/platform
* matrix - https://github.com/matrix-org/gomatrix
* slack - https://github.com/nlopes/slack
* steam - https://github.com/Philipp15b/go-steam
* telegram - https://github.com/go-telegram-bot-api/telegram-bot-api
* xmpp - https://github.com/mattn/go-xmpp
* zulip - https://github.com/ifo/gozulipbot
