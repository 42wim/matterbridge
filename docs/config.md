# Configuring matterbridge

> [!WARNING]
> This page explains the basic concepts for configuring matterbridge, and gives a very simple example.
> For more complete configuration options refer to:
>   - [settings.md](settings.md) for global settings
>   - [protocols/](protocols/) documentation folder for protocol-specific settings
>   - [advanced/](advanced/) documentation folder for advanced settings, such as Tengo scripting and media servers (TODO: rewrite those docs)

## Concepts

matterbridge configuration makes use of a few concepts:

- a `protocol` is a way to reach a certain chat server ; some protocols support multiple servers (eg. IRC/XMPP/Matrix), while others don't (eg. Discord/Telegram/Whatsapp)
- an `account` is a way for your matterbridge bot to reach a specific server (usually with a username/psasword or an API token)
- a `channel` is a chatroom on a specific server
- a `gateway` is a virtual room that aggregates several channels to relay messages between them with your matterbridge bot

### Protocols

Protocols are ways for matterbridge to connect to a network/platform. The list of supported protocols, and supported features for that protocol, is available in [protocols/](protocols/).

### Accounts

An account allows the bot to connect to a certain server using a protocol. Your matterbridge bot may have different accounts (and identities) on the same protocol, for example being connected with two different Discord accounts. In the case of self-hosted centralized protocols such as IRC or Mattermost, it may even be required to have different accounts on different servers in order to reach all the chatrooms you'd like to bridge.

> [!TIP]
> For documentation on how to create an account on specific protocols for your bot, refer to specific protocols docs in [protocols/](protocols/).

A basic account configuration looks like this:

```toml
# Telegram has only one server (no need to specify the address),
# does not have usernames/passwords for bots (only API tokens)
# and does not allow to configure the bot name.
[telegram.myaccount]
Token="MySecretBotToken"
```

As explained, you can have different identities on the same network/protocol:

```toml
# An IRC server (or as IRC people call it, an "IRC network") supports
# authentication by username/password.
[irc.libera]
Server="irc.libera.chat:6667"
Nick="yourliberaaccount"
NickServPassword="yourpassword"
[irc.geeknode]
Server="irc.geeknode.org:6667"
Nick="yourgeeknodeaccount"
NickServPassword="yourpassword"
```

Some protocols allow configuring the bot's name, others don't.

> [!NOTE]
> Some servers (such as Telegram) don't allow the account to have two connections at once. In that case, it's only possible to connect a single matterbridge instance to the account (at a given moment), and running a different bot on the same account is not possible.

### Channels

Channels are specific chatrooms on a certain network. For example, `#matterbridge` on `irc.libera.chat`, or `matterbridge@conference.jabber.de`.
That's where the humans meet and chat, and that's where we want the matterbridge bot to go to relay messages.

Depending on the protocol, channels may be represented by a name (such as `#matterbridge` or `matterbridge` in the previous example), or by a numeric identifier (like `0213809218`).

### Gateways

Gateways are a matterbridge-specific concept. They are not what is known as gateways in [Jabber/XMPP](https://joinjabber.org/tutorials/gateways/) or bridges in [Jabber/XMPP](https://joinjabber.org/tutorials/bridges/) or [Matrix](https://matrix.org/ecosystem/bridges/).

Gateways are simple configuration to link different channels together, and let matterbridge know how to relay messages. They have two basic settings: `name` (for logging), and `enable` (`true` or `false`). Then gateway have an `inout` list of `account`/`channel` to bridge together:

```toml
[[gateway]]
name="mygateway"
enable=true

[[gateway.inout]]
account="irc.libera"
channel="#testing"

[[gateway.inout]]
account="irc.geeknode"
channel="#testingheretoo"
```

The gateway configuration reuses previous configured accounts, adding a specific channel on that account's server.

From this simple example, you can:

- add a new channel to the same bridged discussion, by adding a new `[[gateway.inout]]` section
- add an entirely new discussion bridging other channels, by creating a new `[[gateway]]` section, with the corresponding `[[gateway.inout]]` sections

## Basic configuration

Taking the example from the previous section, a full valid configuration file (except for ommitted bot passwords), would be:

```toml
[irc.libera]
Server="irc.libera.chat:6667"
Nick="yourliberaaccount"
NickServPassword="yourpassword"

[irc.geeknode]
Server="irc.geeknode.org:6667"
Nick="yourgeeknodeaccount"
NickServPassword="yourpassword"

[[gateway]]
name="mygateway"
enable=true

[[gateway.inout]]
account="irc.libera"
channel="#testing"

[[gateway.inout]]
account="irc.geeknode"
channel="#testingheretoo"
```
