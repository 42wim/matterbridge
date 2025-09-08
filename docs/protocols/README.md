# Supported protocols

Matterbridge supports many protocols, although not all of them support all features. Here's a list of officially-maintained and 3rd-party-maintained networks for matterbridge.

## Natively supported

- [Discord](https://discordapp.com)
  - Matterbridge docs:
    - [discord docs](discord/)
    - [discord settings](discord/settings.md)
  - Channel format:
    - by name: `channel_name` (without the leading `#`)
    - by ID: `ID:channel_id`
- [IRC](http://www.mirc.com/servers.html)
  - Matterbridge docs:
    - [irc docs](irc/)
    - [irc settings](irc/settings.md)
  - Channel format: `#channel_name` (it's all lowercase, and don't forget the leading `#`)
- [Jabber](https://joinjabber.org/) is the same as XMPP
  - Matterbridge docs:
    - [xmpp docs](xmpp/)
    - [xmpp settings](xmpp/settings.md)
  - Channel format: `channel_name` (for `channel_name@muc.server.org` where `muc.server.org` has been configured as `Muc` for the corresponding xmpp account)
- [Keybase](https://keybase.io)
  - Matterbridge [keybase docs](keybase/)
  - Channel format: `channel_name` (without the leading `#`), or `general` if your team doesn't have multiple channels
- [Matrix](https://matrix.org)
  - Matterbridge docs:
    - [matrix docs](matrix/)
    - [matrix settings](matrix/settings.md)
  - Channel format: `#channel_name:server.org`
- [Mattermost](https://github.com/mattermost/mattermost-server/)
  - Matterbridge docs:
    - [matterbridge docs](matterbridge/)
    - [matterbridge settings](matterbridge/settings.md)
  - Channel format:
    - by name: `channel_name` as seen in the URL `https://yourmattermostserver/yourteam/channels/channel_name`
    - by ID: `ID:channel_id`
- [Microsoft Teams](https://teams.microsoft.com)
  - Matterbridge [msteams docs](msteams/)
  - Channel format: `19:82caxx@thread.skype` as seen in the URL `?threadId=19:82caxx@thread.skype`
- [Mumble](https://www.mumble.info/)
  - Matterbridge [mumble docs](mumble/)
  - Channel format: `channel_id` as seen in the channel's `Edit` window
- [Nextcloud Talk](https://nextcloud.com/talk/)
  - Matterbridge [nctalk docs](nctalk/)
  - Channel format: `channel_id` as seen at the end of URL (eg. `xs25tz5y`)
- [Rocket.chat](https://rocket.chat)
  - Matterbridge docs:
    - [rocketchat docs](rocketchat/)
    - [rocketchat settings](rocketchat/settings.md)
  - Channel format: `#channel_name` (don't forget the leading `#`, even on private channels)
- [Slack](https://slack.com)
  - Matterbridge docs:
    - [slack docs](slack/)
    - [slack settings](slack/settings.md)
  - Channel format:
    - by name: `channel_name` (without the leading `#`)
    - by ID: `ID:channel_id` (does not work with webhooks!)
- [Ssh-chat](https://github.com/shazow/ssh-chat)
  - Matterbridge [sshchat docs](sshchat/)
  - Channel format: Only a single `sshchat` channel is supported
- [Telegram](https://telegram.org)
  - Matterbridge docs:
    - [telegram docs](telegram/)
    - [telegram settings](telegram/settings.md)
  - Channel format:
    - for channels/groups: `-channel_id` where `channel_id` is a large number (see FAQ)
    - for forum topics (sub-groups): `-100channel_id/topic_id` (see FAQ), except the first `General` topic which is `-100channel_id` (**not** `-100channel_id/1`)
- [Twitch](https://twitch.tv)
  - Matterbridge [twitch docs](twitch/)
  - Channel format: `#channel_name` (it's all lowercase, and don't forget the leading `#`)
- [VK](https://vk.com/)
  - Matterbridge [vk docs](vk/)
  - Channel format: `channel_id` (see FAQ)
- [WhatsApp](https://www.whatsapp.com/)
  - Matterbridge docs:
    - [whatsapp docs](whatsapp/)
    - [whatsapp settings](whatsapp/settings.md)
  - Channel format:
    - by JID: `channel_id@g.us` (if `Channel=""`, matterbridge will list all channels known to the bot)
    - by name: `channel_name` (**not recommended**, and matterbridge will warn you against it because group names can change over time)
- [XMPP](https://xmpp.org)
  - Matterbridge docs:
    - [xmpp docs](xmpp/)
    - [xmpp settings](xmpp/settings.md)
  - Channel format: `channel_name` (for `channel_name@muc.server.org` where `muc.server.org` has been configured as `Muc` for the corresponding xmpp account)
- [Zulip](https://zulipchat.com)
  - Matterbridge docs:
    - [zulip docs](zulip/)
    - [zulip settings](zulip/settings.md)
  - Channel format: `stream/topic:channel_name` (where `channel_name` has no leading `#`)

## Dropped official support

- [Gitter](https://gitter.im)
  - Has moved to matrix protocol
- [Harmony](https://harmonyapp.io)
  - Does not exist anymore?

- [Steam](https://store.steampowered.com/)
  - Not supported anymore, see [here](https://github.com/Philipp15b/go-steam/issues/94) for more info.

## 3rd party via matterbridge api

- [Delta Chat](https://github.com/deltachat-bot/matterdelta)
- [Minecraft](https://github.com/raws/mattercraft)
- [Minecraft](https://gitlab.com/Programie/MatterBukkit)
- [MatterLink](https://github.com/elytra/MatterLink) (Matterbridge link for Minecraft Forge server chat, archived)
- [MatterCraft](https://github.com/raws/mattercraft) (Matterbridge link for Minecraft Forge server chat)
- [MatterBukkit](https://gitlab.com/Programie/MatterBukkit) (Matterbridge link for Minecraft Bukkit/Spigot server chat)
- [pyCord](https://github.com/NikkyAI/pyCord) (crossplatform chatbot)
- [Mattereddit](https://github.com/bonehurtingjuice/mattereddit) (Reddit chat support)
- [ServUO-matterbridge](https://github.com/kuoushi/ServUO-Matterbridge) (A matterbridge connector for ServUO servers)
- [beerchat](https://github.com/mt-mods/beerchat) (Matterbridge link for minetest)
- [nextcloud talk](https://github.com/nextcloud/talk_matterbridge) (Integrates matterbridge in Nextcloud Talk)


### Past 3rd party projects

- [Discourse](https://github.com/DeclanHoare/matterbabble)
- [Facebook messenger](https://github.com/powerjungle/fbridge-asyncio)
- [Facebook messenger](https://github.com/VictorNine/fbridge)
- [Minecraft](https://github.com/elytra/MatterLink)
- [Reddit](https://github.com/bonehurtingjuice/mattereddit)
- [MatterAMXX](https://github.com/andrewlindberg/MatterAMXX): [Counter-Strike, half-life and more](https://forums.alliedmods.net/showthread.php?t=319430)
- [Vintage Story](https://github.com/NikkyAI/vs-matterbridge) (last commit: February 4th 2021)
- [Ultima Online Emulator](https://github.com/kuoushi/ServUO-Matterbridge)
- [Teamspeak](https://github.com/Archeb/ts-matterbridge)
- [ts-matterbridge](https://github.com/Archeb/ts-matterbridge) (Integrate teamspeak chat with matterbridge) (archived September 25th 2022)
