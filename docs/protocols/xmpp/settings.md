# XMPP settings

> [!TIP]
> This page contains the details about xmpp settings. More general information about xmpp support in matterbridge can be found in [README.md](README.md).

> [!NOTE]
> XMPP (the protocol) is also known as Jabber (the open federation). These
> two terms are used interchangeably. To learn more about Jabber/XMPP,
> see [joinjabber.org](https://joinjabber.org/).

## Jid

Jabber Identifier, the XMPP login for matterbridge's account.

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Jid="user@example.com"
  ```

## MUC

The Multi User Chat (MUC) server where the bot will find the defined gateway
channels. At the moment, bridging a room on a different MUC requires creating
a separate account entry in the configuration.

TODO: test if a matterbridge instance can be connected to the same account
      with two configurations at the same time; this is allowed by XMPP
      protocol but requires matterbridge to behave properly in terms
      of XMPP protocol

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Muc="conference.jabber.example.com"
  ```

## Nick 

Your nick in the rooms

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Nick="xmppbot"
  ```

## NoTLS

Enable this to make an insecure plaintext connection to your xmpp server.
This is usually not permitted by XMPP servers even on localhost.

- Setting: **OPTIONAL**
- Format: *boolean*
- Example:
  ```toml
  NoTLS=true
  ```

## Password

Password for the Jid's account.

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Password="yourpass"
  ```

## Server

XMPP server to connect to.

- Setting: **REQUIRED**
- Format: *string* (hostname:port)
- Example:
  ```toml
  Server="jabber.example.com:5222"
  ```
