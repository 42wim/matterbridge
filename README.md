<div align="center">

# matterbridge

![Matterbridge Logo](img/matterbridge-notext.gif)<br />
**A simple chat bridge**<br />
Letting people be where they want to be with the magic of [interoperability](https://en.wikipedia.org/wiki/Interoperability).<br />

---

[![Download stable](https://img.shields.io/github/release/42wim/matterbridge.svg?label=download%20stable)](https://github.com/42wim/matterbridge/releases/latest)

---

</div>

matterbridge is a solution to connect users on different platforms/protocols, allowing them to chat in the best conditions possible. Despite the name, Matter<em>most</em> isn't required to run matter<em>bridge</em>.

## Features

Many features are available, but not all of them are supported on all protocols:

- [x] Bridge many rooms from supported protocols
- [x] Message threads/topics, and replies
- [x] Attachments, file uploads, and inline images
- [x] Transparent bridging with spoofed usernames and avatars
- [x] Private groups

**The complete and up-to-date list of supported protocols is in [docs/protocols/](docs/protocols/).** Additionally, we have an [API for 3rd party integration](https://github.com/42wim/matterbridge/wiki/Features#api) if you'd like to add a custom bridge without implementing it in this codebase.

![Screenshot of users discussing from different networks using matterbridge](https://user-images.githubusercontent.com/849975/52647227-9c3a5300-2ee4-11e9-9c57-ea096473aba8.png)

## Getting started with matterbridge

Get matterbridge up-and-running in a few minutes in 3 simple steps:

- [Setting up matterbridge](docs/setup.md) (see [docs/compiling.md](docs/compiling.md) for compiling from source)
- [Configuring matterbridge](docs/config.md)
- [Running matterbridge](docs/running.md) (CLI or as a systemd service)

## Documentation

See [docs/](docs/) folder in this repository.

## Contributing

This project is licensed under the Apache License 2.0. A copy can be found in the [LICENSE](LICENSE) file in this repository.

You are welcome to submit pull requests, report bugs and request new features. matterbridge is a volunteer-run project and you are expected to behave with respect for the maintainers and other users. In particular, harassment and hate speech are not welcome.

For more development guidelines, see [docs/development/](docs/development/).

### Chat with us

Questions or want to test on your favorite platform? Join us on:

- federated networks: [Gitter][mb-gitter], [Jabber/XMPP][mb-xmpp], [Matrix][mb-matrix]
- non-free centralized networks: [Discord][mb-discord], [Keybase][mb-keybase], [Slack][mb-slack], [Telegram][mb-telegram], [Twitch][mb-twitch]
- self-hostable centralized networks: [IRC][mb-irc], [Mattermost][mb-mattermost], [Rocket.Chat][mb-rocketchat], [Zulip][mb-zulip]

## Related projects

- [jwflory/ansible-role-matterbridge](https://galaxy.ansible.com/jwflory/matterbridge) (Ansible role to simplify deploying Matterbridge)
- [matterbridge autoconfig](https://github.com/patcon/matterbridge-autoconfig)
- [matterbridge config viewer](https://github.com/patcon/matterbridge-heroku-viewer)
- [matterbridge-heroku](https://github.com/cadecairos/matterbridge-heroku)
- [mattermost-plugin](https://github.com/matterbridge/mattermost-plugin) - Run matterbridge as a plugin in mattermost
- [isla](https://github.com/alphachung/isla) (Bot for Discord-Telegram groups used alongside matterbridge)

## Thanks

Matterbridge wouldn't exist without amazing libraries, without [@42wim](https://github.com/42wim) who started the project, and without the 100+ contributors who participated in this adventure. See [docs/credits.md](docs/credits.md) for more complete credits.

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
[mb-xmpp]: xmpp:matterbridge@conference.jabber.de?join
[mb-zulip]: https://matterbridge.zulipchat.com/register/

