<div align="center">

# matterbridge

![Matterbridge Logo](img/matterbridge-notext.gif)<br />
**A simple chat bridge**<br />
Letting people be where they want to be.<br />
<sub>Bridges between a growing number of protocols. Click below to demo or join the development chat.</sub>

   <sup>

[Discord][mb-discord] |
[Gitter][mb-gitter] |
[IRC][mb-irc] |
[Keybase][mb-keybase] |
[Matrix][mb-matrix] |
[Mattermost][mb-mattermost] |
[MSTeams][mb-msteams] |
[Rocket.Chat][mb-rocketchat] |
[Slack][mb-slack] |
[Telegram][mb-telegram] |
[Twitch][mb-twitch] |
[WhatsApp][mb-whatsapp] |
[XMPP][mb-xmpp] |
[Zulip][mb-zulip] |
And more...
</sup>

---

[![Download stable](https://img.shields.io/github/release/42wim/matterbridge.svg?label=download%20stable)](https://github.com/42wim/matterbridge/releases/latest)
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

# Table of Contents

- [matterbridge](#matterbridge)
- [Table of Contents](#table-of-contents)
  - [Features](#features)
    - [Natively supported](#natively-supported)
    - [3rd party via matterbridge api](#3rd-party-via-matterbridge-api)
    - [API](#api)
  - [Chat with us](#chat-with-us)
  - [Screenshots](#screenshots)
  - [Installing / upgrading](#installing--upgrading)
    - [Binaries](#binaries)
    - [Packages](#packages)
  - [Building](#building)
  - [Building with whatsapp (beta) multidevice support](#building-with-whatsapp-beta-multidevice-support)
  - [Configuration](#configuration)
    - [Basic configuration](#basic-configuration)
    - [Settings](#settings)
    - [Advanced configuration](#advanced-configuration)
      - [Examples](#examples)
        - [Bridge mattermost (off-topic) - irc (#testing)](#bridge-mattermost-off-topic---irc-testing)
        - [Bridge slack (#general) - discord (general)](#bridge-slack-general---discord-general)
    - [Running](#running)
      - [Docker](#docker)
      - [Systemd](#systemd)
    - [Changelog](#changelog)
    - [FAQ](#faq)
    - [Related projects](#related-projects)
    - [Articles / Tutorials](#articles--tutorials)
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

- [Discord](https://discordapp.com)
- [Gitter](https://gitter.im)
- [Harmony](https://harmonyapp.io)
- [IRC](http://www.mirc.com/servers.html)
- [Keybase](https://keybase.io)
- [Matrix](https://matrix.org)
- [Mattermost](https://github.com/mattermost/mattermost-server/)
- [Microsoft Teams](https://teams.microsoft.com)
- [Mumble](https://www.mumble.info/)
- [Nextcloud Talk](https://nextcloud.com/talk/)
- [Rocket.chat](https://rocket.chat)
- [Slack](https://slack.com)
- [Ssh-chat](https://github.com/shazow/ssh-chat)
- ~~[Steam](https://store.steampowered.com/)~~
  - Not supported anymore, see [here](https://github.com/Philipp15b/go-steam/issues/94) for more info.
- [Telegram](https://telegram.org)
- [Twitch](https://twitch.tv)
- [VK](https://vk.com/)
- [WhatsApp](https://www.whatsapp.com/)
  - Whatsapp legacy is natively supported
  - Whatsapp multidevice beta is natively supported but you need to build yourself, see [here](#building-with-whatsapp-beta-multidevice-support)
- [XMPP](https://xmpp.org)
- [Zulip](https://zulipchat.com)

### 3rd party via matterbridge api

- [Delta Chat](https://github.com/deltachat-bot/matterdelta)
- [Discourse](https://github.com/DeclanHoare/matterbabble)
- [Facebook messenger](https://github.com/powerjungle/fbridge-asyncio)
- [Facebook messenger](https://github.com/VictorNine/fbridge)
- [Minecraft](https://github.com/elytra/MatterLink)
- [Minecraft](https://github.com/raws/mattercraft)
- [Minecraft](https://gitlab.com/Programie/MatterBukkit)
- [Reddit](https://github.com/bonehurtingjuice/mattereddit)
- [Counter-Strike, half-life and more](https://forums.alliedmods.net/showthread.php?t=319430)
- [MatterAMXX](https://github.com/GabeIggy/MatterAMXX)
- [Vintage Story](https://github.com/NikkyAI/vs-matterbridge)
- [Ultima Online Emulator](https://github.com/kuoushi/ServUO-Matterbridge)
- [Teamspeak](https://github.com/Archeb/ts-matterbridge)

### API

The API is basic at the moment.
More info and examples on the [wiki](https://github.com/42wim/matterbridge/wiki/Api).

Used by the projects below. Feel free to make a PR to add your project to this list.

- [MatterLink](https://github.com/elytra/MatterLink) (Matterbridge link for Minecraft Forge server chat, archived)
- [MatterCraft](https://github.com/raws/mattercraft) (Matterbridge link for Minecraft Forge server chat)
- [MatterBukkit](https://gitlab.com/Programie/MatterBukkit) (Matterbridge link for Minecraft Bukkit/Spigot server chat)
- [pyCord](https://github.com/NikkyAI/pyCord) (crossplatform chatbot)
- [Mattereddit](https://github.com/bonehurtingjuice/mattereddit) (Reddit chat support)
- [fbridge-asyncio](https://github.com/powerjungle/fbridge-asyncio) (Facebook messenger support)
- [fbridge](https://github.com/VictorNine/fbridge) (Facebook messenger support)
- [matterbabble](https://github.com/DeclanHoare/matterbabble) (Discourse support)
- [MatterAMXX](https://forums.alliedmods.net/showthread.php?t=319430) (Counter-Strike, half-life and more via AMXX mod)
- [Vintage Story](https://github.com/NikkyAI/vs-matterbridge)
- [ServUO-matterbridge](https://github.com/kuoushi/ServUO-Matterbridge) (A matterbridge connector for ServUO servers)
- [ts-matterbridge](https://github.com/Archeb/ts-matterbridge) (Integrate teamspeak chat with matterbridge)
- [beerchat](https://github.com/mt-mods/beerchat) (Matterbridge link for minetest)

## Chat with us

Questions or want to test on your favorite platform? Join below:

- [Discord][mb-discord]
- [Gitter][mb-gitter]
- [IRC][mb-irc]
- [Keybase][mb-keybase]
- [Matrix][mb-matrix]
- [Mattermost][mb-mattermost]
- [Rocket.Chat][mb-rocketchat]
- [Slack][mb-slack]
- [Telegram][mb-telegram]
- [Twitch][mb-twitch]
- [XMPP][mb-xmpp] (matterbridge@conference.jabber.de)
- [Zulip][mb-zulip]

## Screenshots

See <https://github.com/42wim/matterbridge/wiki>

## Installing / upgrading

### Binaries

- Latest stable release [v1.26.0](https://github.com/42wim/matterbridge/releases/latest)
- Development releases (follows master) can be downloaded [here](https://github.com/42wim/matterbridge/actions) selecting the latest green build and then artifacts.

To install or upgrade just download the latest [binary](https://github.com/42wim/matterbridge/releases/latest). On \*nix platforms you may need to make the binary executable - you can do this by running `chmod a+x` on the binary (example: `chmod a+x matterbridge-1.24.1-linux-64bit`). After downloading (and making the binary executable, if necessary), follow the instructions on the [howto](https://github.com/42wim/matterbridge/wiki/How-to-create-your-config) for a step by step walkthrough for creating your configuration.

### Packages

- [Overview](https://repology.org/metapackage/matterbridge/versions)
- [snap](https://snapcraft.io/matterbridge)
- [scoop](https://github.com/42wim/scoop-bucket)

## Building

Most people just want to use binaries, you can find those [here](https://github.com/42wim/matterbridge/releases/latest)

If you really want to build from source, follow these instructions:
Go 1.18+ is required. Make sure you have [Go](https://golang.org/doc/install) properly installed.

Building the binary with **all** the bridges enabled needs about 3GB RAM to compile.
You can reduce this memory requirement to 0,5GB RAM by adding the `nomsteams` tag if you don't need/use the Microsoft Teams bridge.

Matterbridge can be build without gcc/c-compiler: If you're running on windows first run `set CGO_ENABLED=0` on other platforms you prepend `CGO_ENABLED=0` to the `go build` command. (eg `CGO_ENABLED=0 go install github.com/42wim/matterbridge`)

To install the latest stable run:

```bash
go install github.com/42wim/matterbridge
```

To install the latest dev run:

```bash
go install github.com/42wim/matterbridge@master
```

To install the latest stable run without msteams or zulip bridge:

```bash
go install -tags nomsteams,nozulip github.com/42wim/matterbridge
```

You should now have matterbridge binary in the ~/go/bin directory:

```bash
$ ls ~/go/bin/
matterbridge
```

## Building with whatsapp (beta) multidevice support

Because the library we use for Whatsapp multidevice support includes a GPL3 library we can not provide you binaries.
(as this would require the Matterbridge to change it license to GPL)

Matterbridge can be build without gcc/c-compiler: If you're running on windows first run `set CGO_ENABLED=0` on other platforms you prepend `CGO_ENABLED=0` to the `go build` command. (eg `CGO_ENABLED=0 go install github.com/42wim/matterbridge`)

So this means you have to build it yourself using the instructions below:

```bash
go install -tags whatsappmulti github.com/42wim/matterbridge@master
```

If you're low on memory and don't need msteams:

```bash
go install -tags nomsteams,whatsappmulti github.com/42wim/matterbridge@master
```

You should now have matterbridge binary in the ~/go/bin directory:

```bash
$ ls ~/go/bin/
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
    [irc.libera]
    Server="irc.libera.chat:6667"
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
    account="irc.libera"
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

```bash
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

Please take a look at the [Docker Wiki page](https://github.com/42wim/matterbridge/wiki/Deploy:-Docker) for more information.

### Systemd

Please take a look at the [Service Files page](https://github.com/42wim/matterbridge/wiki/Service-files) for more information.

## Changelog

See [changelog.md](https://github.com/42wim/matterbridge/blob/master/changelog.md)

## FAQ

See [FAQ](https://github.com/42wim/matterbridge/wiki/FAQ)

## Related projects

- [jwflory/ansible-role-matterbridge](https://galaxy.ansible.com/jwflory/matterbridge) (Ansible role to simplify deploying Matterbridge)
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
- [nextcloud talk](https://github.com/nextcloud/talk_matterbridge) (Integrates matterbridge in Nextcloud Talk)
- [mattercraft](https://github.com/raws/mattercraft) (Minecraft bridge)
- [vs-matterbridge](https://github.com/NikkyAI/vs-matterbridge) (Vintage Story bridge)
- [ServUO-matterbridge](https://github.com/kuoushi/ServUO-Matterbridge) (A matterbridge connector for ServUO servers)
- [ts-matterbridge](https://github.com/Archeb/ts-matterbridge) (Integrate teamspeak chat with matterbridge)

## Articles / Tutorials

- [matterbridge on kubernetes](https://medium.freecodecamp.org/using-kubernetes-to-deploy-a-chat-gateway-or-when-technology-works-like-its-supposed-to-a169a8cd69a3)
- <https://mattermost.com/blog/connect-irc-to-mattermost/>
- <https://blog.valvin.fr/2016/09/17/mattermost-et-un-channel-irc-cest-possible/>
- <https://blog.brightscout.com/top-10-mattermost-integrations/>
- <https://www.algoo.fr/blog/2018/01/19/recouvrez-votre-liberte-en-quittant-slack-pour-un-mattermost-auto-heberge/>
- <https://kopano.com/blog/matterbridge-bridging-mattermost-chat/>
- <https://www.stitcher.com/s/?eid=52382713>
- <https://daniele.tech/2019/02/how-to-use-matterbridge-to-connect-2-different-slack-workspaces/>
- <https://userlinux.net/mattermost-and-matterbridge.html>
- <https://nextcloud.com/blog/bridging-chat-services-in-talk/>
- <https://minecraftchest1.wordpress.com/2021/06/05/how-to-install-and-setup-matterbridge/>
- Youtube: [whatsapp - telegram bridging](https://www.youtube.com/watch?v=W-VXISoKtNc)

## Thanks

<p>This project is supported by:</p>
<p>
  <a href="https://www.digitalocean.com/">
    <img src="https://opensource.nyc3.cdn.digitaloceanspaces.com/attribution/assets/SVG/DO_Logo_horizontal_blue.svg" width="201px">
  </a>
</p>

Matterbridge wouldn't exist without these libraries:

- discord - <https://github.com/bwmarrin/discordgo>
- echo - <https://github.com/labstack/echo>
- gops - <https://github.com/google/gops>
- gozulipbot - <https://github.com/ifo/gozulipbot>
- gumble - <https://github.com/layeh/gumble>
- harmony - <https://github.com/harmony-development/shibshib>
- irc - <https://github.com/lrstanley/girc>
- keybase - <https://github.com/keybase/go-keybase-chat-bot>
- matrix - <https://github.com/matrix-org/gomatrix>
- mattermost - <https://github.com/mattermost/mattermost-server>
- msgraph.go - <https://github.com/yaegashi/msgraph.go>
- mumble - <https://github.com/layeh/gumble>
- nctalk - <https://github.com/gary-kim/go-nc-talk>
- rocketchat - <https://github.com/RocketChat/Rocket.Chat.Go.SDK>
- slack - <https://github.com/nlopes/slack>
- sshchat - <https://github.com/shazow/ssh-chat>
- steam - <https://github.com/Philipp15b/go-steam>
- telegram - <https://github.com/go-telegram-bot-api/telegram-bot-api>
- tengo - <https://github.com/d5/tengo>
- vk - <https://github.com/SevereCloud/vksdk>
- whatsapp - <https://github.com/Rhymen/go-whatsapp>
- whatsapp - <https://github.com/tulir/whatsmeow>
- xmpp - <https://github.com/mattn/go-xmpp>
- zulip - <https://github.com/ifo/gozulipbot>

<!-- Links -->

[mb-discord]: https://discord.gg/AkKPtrQ
[mb-gitter]: https://gitter.im/42wim/matterbridge
[mb-irc]: https://web.libera.chat/#matterbridge
[mb-keybase]: https://keybase.io/team/matterbridge
[mb-matrix]: https://riot.im/app/#/room/#matterbridge:matrix.org
[mb-mattermost]: https://framateam.org/signup_user_complete/?id=tfqm33ggop8x3qgu4boeieta6e
[mb-msteams]: https://teams.microsoft.com/join/hj92x75gd3y7
[mb-rocketchat]: https://open.rocket.chat/channel/matterbridge
[mb-slack]: https://join.slack.com/t/matterbridgechat/shared_invite/zt-2ourq2h2-7YvyYBq2WFGC~~zEzA68_Q
[mb-telegram]: https://t.me/Matterbridge
[mb-twitch]: https://www.twitch.tv/matterbridge
[mb-whatsapp]: https://www.whatsapp.com/
[mb-xmpp]: https://inverse.chat/
[mb-zulip]: https://matterbridge.zulipchat.com/register/
