<div align="center">

# matterbridge

![Matterbridge Logo](img/matterbridge-notext.gif)<br />
**A simple chat bridge**<br />
Letting people be where they want to be.<br />
<sub>Bridges between a growing number of protocols. Click below to demo or join the development chat.</sub>

   <sup>

[Gitter][mb-gitter] |
[IRC][mb-irc] |
[Discord][mb-discord] |
[Matrix][mb-matrix] |
[Slack][mb-slack] |
[Mattermost][mb-mattermost] |
[Rocket.Chat][mb-rocketchat] |
[XMPP][mb-xmpp] |
[Twitch][mb-twitch] |
[WhatsApp][mb-whatsapp] |
[Zulip][mb-zulip] |
[Telegram][mb-telegram] |
[Keybase][mb-keybase] |
And more...
</sup>

---

[![Download stable](https://img.shields.io/github/release/42wim/matterbridge.svg?label=download%20stable)](https://github.com/42wim/matterbridge/releases/latest)
[![Download dev](https://img.shields.io/bintray/v/42wim/nightly/Matterbridge.svg?label=download%20dev&colorB=007ec6)](https://bintray.com/42wim/nightly/Matterbridge/_latestVersion)
[![Maintainability](https://api.codeclimate.com/v1/badges/82dff70ef2ba85a6173a/maintainability)](https://codeclimate.com/github/42wim/matterbridge/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/82dff70ef2ba85a6173a/test_coverage)](https://codeclimate.com/github/42wim/matterbridge/test_coverage)<br />

  <hr />
</div>
<div align="right"><sup>

**Note:** Matter<em>most</em> isn't required to run matter<em>bridge</em>.</sup></div>

<p>
  <a href="https://www.digitalocean.com/">
    <img src="https://opensource.nyc3.cdn.digitaloceanspaces.com/attribution/assets/PoweredByDO/DO_Powered_by_Badge_blue.svg" width="201px">
  </a>
</p>

### Table of Contents

- [Features](https://github.com/42wim/matterbridge/wiki/Features)
  - [Natively supported](#natively-supported)
  - [3rd party via matterbridge api](#3rd-party-via-matterbridge-api)
  - [API](#API)
- [Chat with us](#chat-with-us)
- [Screenshots](https://github.com/42wim/matterbridge/wiki/)
- [Installing/upgrading](#installing--upgrading)
  - [Binaries](#binaries)
- [Building](#building)
- [Configuration](#configuration)
  - [Howto](https://github.com/42wim/matterbridge/wiki/How-to-create-your-config)
  - [Settings](#settings)
  - [Examples](#examples)
- [Running](#running)
  - [Docker](#docker)
- [Changelog](#changelog)
- [FAQ](#faq)
- [Related projects](#related-projects)
- [Articles](#articles)
- [Thanks](#thanks)

## Features

- [Support bridging between any protocols](https://github.com/42wim/matterbridge/wiki/Features#support-bridging-between-any-protocols)
- [Support multiple gateways(bridges) for your protocols](https://github.com/42wim/matterbridge/wiki/Features#support-multiple-gatewaysbridges-for-your-protocols)
- [Message edits and deletes](https://github.com/42wim/matterbridge/wiki/Features#message-edits-and-deletes)
- Preserves threading when possible
- [Attachment / files handling](https://github.com/42wim/matterbridge/wiki/Features#attachment--files-handling)
- [Username and avatar spoofing](https://github.com/42wim/matterbridge/wiki/Features#username-and-avatar-spoofing)
- [Private groups](https://github.com/42wim/matterbridge/wiki/Features#private-groups)
- [API](https://github.com/42wim/matterbridge/wiki/Features#api)

### Natively supported

- [Mattermost](https://github.com/mattermost/mattermost-server/) 4.x, 5.x
- [IRC](http://www.mirc.com/servers.html)
- [XMPP](https://xmpp.org)
- [Gitter](https://gitter.im)
- [Slack](https://slack.com)
- [Discord](https://discordapp.com)
- [Telegram](https://telegram.org)
- [Rocket.chat](https://rocket.chat)
- [Matrix](https://matrix.org)
- [Steam](https://store.steampowered.com/)
- [Twitch](https://twitch.tv)
- [Ssh-chat](https://github.com/shazow/ssh-chat)
- [WhatsApp](https://www.whatsapp.com/)
- [Zulip](https://zulipchat.com)
- [Keybase](https://keybase.io)

### 3rd party via matterbridge api

- [Minecraft](https://github.com/elytra/MatterLink)
- [Reddit](https://github.com/bonehurtingjuice/mattereddit)
- [Facebook messenger](https://github.com/VictorNine/fbridge)
- [Discourse](https://github.com/DeclanHoare/matterbabble)

### API

The API is basic at the moment.
More info and examples on the [wiki](https://github.com/42wim/matterbridge/wiki/Api).

Used by the projects below. Feel free to make a PR to add your project to this list.

- [MatterLink](https://github.com/elytra/MatterLink) (Matterbridge link for Minecraft Server chat)
- [pyCord](https://github.com/NikkyAI/pyCord) (crossplatform chatbot)
- [Mattereddit](https://github.com/bonehurtingjuice/mattereddit) (Reddit chat support)
- [fbridge](https://github.com/VictorNine/fbridge) (Facebook messenger support)
- [matterbabble](https://github.com/DeclanHoare/matterbabble) (Discourse support)

## Chat with us

Questions or want to test on your favorite platform? Join below:

- [Gitter][mb-gitter]
- [IRC][mb-irc]
- [Discord][mb-discord]
- [Matrix][mb-matrix]
- [Slack][mb-slack]
- [Mattermost][mb-mattermost]
- [Rocket.Chat][mb-rocketchat]
- [XMPP][mb-xmpp] (matterbridge@conference.jabber.de)
- [Twitch][mb-twitch]
- [Zulip][mb-zulip]
- [Telegram][mb-telegram]

## Screenshots

See https://github.com/42wim/matterbridge/wiki

## Installing / upgrading

### Binaries

- Latest stable release [v1.16.1](https://github.com/42wim/matterbridge/releases/latest)
- Development releases (follows master) can be downloaded [here](https://dl.bintray.com/42wim/nightly/)

To install or upgrade just download the latest [binary](https://github.com/42wim/matterbridge/releases/latest) and follow the instructions on the [howto](https://github.com/42wim/matterbridge/wiki/How-to-create-your-config) for a step by step walkthrough for creating your configuration.

### Packages

- [Overview](https://repology.org/metapackage/matterbridge/versions)

## Building

Most people just want to use binaries, you can find those [here](https://github.com/42wim/matterbridge/releases/latest)

If you really want to build from source, follow these instructions:
Go 1.9+ is required. Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH](https://golang.org/doc/code.html#GOPATH).

After Go is setup, download matterbridge to your \$GOPATH directory.

```
cd $GOPATH
go get github.com/42wim/matterbridge
```

You should now have matterbridge binary in the bin directory:

```
$ ls bin/
matterbridge
```

## Configuration

### Basic configuration

See [howto](https://github.com/42wim/matterbridge/wiki/How-to-create-your-config) for a step by step walkthrough for creating your configuration.

### Settings

All possible [settings](https://github.com/42wim/matterbridge/wiki/Settings) for each bridge.

### Advanced configuration

- [matterbridge.toml.sample](https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample) for documentation and an example.

### Examples

#### Bridge mattermost (off-topic) - irc (#testing)

```toml
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

#### Bridge slack (#general) - discord (general)

```toml
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

## Running

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

### Docker

Create your matterbridge.toml file locally eg in `/tmp/matterbridge.toml`

```
docker run -ti -v /tmp/matterbridge.toml:/matterbridge.toml 42wim/matterbridge
```

## Changelog

See [changelog.md](https://github.com/42wim/matterbridge/blob/master/changelog.md)

## FAQ

See [FAQ](https://github.com/42wim/matterbridge/wiki/FAQ)

## Related projects

- [FOSSRIT/infrastructure - roles/matterbridge](https://github.com/FOSSRIT/infrastructure/tree/master/roles/matterbridge) (Ansible role used to automate deployments of Matterbridge)
- [matterbridge autoconfig](https://github.com/patcon/matterbridge-autoconfig)
- [matterbridge config viewer](https://github.com/patcon/matterbridge-heroku-viewer)
- [matterbridge-heroku](https://github.com/cadecairos/matterbridge-heroku)
- [mattereddit](https://github.com/bonehurtingjuice/mattereddit)
- [matterlink](https://github.com/elytra/MatterLink)
- [mattermost-plugin](https://github.com/matterbridge/mattermost-plugin) - Run matterbridge as a plugin in mattermost
- [pyCord](https://github.com/NikkyAI/pyCord) (crossplatform chatbot)
- [fbridge](https://github.com/VictorNine/fbridge) (Facebook messenger support)
- [isla](https://github.com/alphachung/isla) (Bot for Discord-Telegram groups used alongside matterbridge)
- [matterbabble](https://github.com/DeclanHoare/matterbabble) (Connect Discourse threads to Matterbridge)

## Articles

- [matterbridge on kubernetes](https://medium.freecodecamp.org/using-kubernetes-to-deploy-a-chat-gateway-or-when-technology-works-like-its-supposed-to-a169a8cd69a3)
- https://mattermost.com/blog/connect-irc-to-mattermost/
- https://blog.valvin.fr/2016/09/17/mattermost-et-un-channel-irc-cest-possible/
- https://blog.brightscout.com/top-10-mattermost-integrations/
- http://bencey.co.nz/2018/09/17/bridge/
- https://www.algoo.fr/blog/2018/01/19/recouvrez-votre-liberte-en-quittant-slack-pour-un-mattermost-auto-heberge/
- https://kopano.com/blog/matterbridge-bridging-mattermost-chat/
- https://www.stitcher.com/s/?eid=52382713
- https://daniele.tech/2019/02/how-to-use-matterbridge-to-connect-2-different-slack-workspaces/

## Thanks

<p>This project is supported by:</p>
<p>
  <a href="https://www.digitalocean.com/">
    <img src="https://opensource.nyc3.cdn.digitaloceanspaces.com/attribution/assets/SVG/DO_Logo_horizontal_blue.svg" width="201px">
  </a>
</p>

Matterbridge wouldn't exist without these libraries:

- discord - https://github.com/bwmarrin/discordgo
- echo - https://github.com/labstack/echo
- gitter - https://github.com/sromku/go-gitter
- gops - https://github.com/google/gops
- gozulipbot - https://github.com/ifo/gozulipbot
- irc - https://github.com/lrstanley/girc
- mattermost - https://github.com/mattermost/mattermost-server
- matrix - https://github.com/matrix-org/gomatrix
- sshchat - https://github.com/shazow/ssh-chat
- slack - https://github.com/nlopes/slack
- steam - https://github.com/Philipp15b/go-steam
- telegram - https://github.com/go-telegram-bot-api/telegram-bot-api
- xmpp - https://github.com/mattn/go-xmpp
- whatsapp - https://github.com/Rhymen/go-whatsapp/
- zulip - https://github.com/ifo/gozulipbot
- tengo - https://github.com/d5/tengo
- keybase - https://github.com/keybase/go-keybase-chat-bot

<!-- Links -->

[mb-gitter]: https://gitter.im/42wim/matterbridge
[mb-irc]: https://webchat.freenode.net/?channels=matterbridgechat
[mb-discord]: https://discord.gg/AkKPtrQ
[mb-matrix]: https://riot.im/app/#/room/#matterbridge:matrix.org
[mb-slack]: https://join.slack.com/matterbridgechat/shared_invite/MjEwODMxNjU1NDMwLTE0OTk2MTU3NTMtMzZkZmRiNDZhOA
[mb-mattermost]: https://framateam.org/signup_user_complete/?id=tfqm33ggop8x3qgu4boeieta6e
[mb-rocketchat]: https://open.rocket.chat/channel/matterbridge
[mb-xmpp]: https://inverse.chat/
[mb-twitch]: https://www.twitch.tv/matterbridge
[mb-whatsapp]: https://www.whatsapp.com/
[mb-keybase]: https://keybase.io
[mb-zulip]: https://matterbridge.zulipchat.com/register/
[mb-telegram]: https://t.me/Matterbridge
