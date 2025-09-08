# XMPP

- Status: Working
- Maintainers: @poVoq, @selfhoster1312
- Features:
  - attachments: incoming/outgoing

> [!NOTE]
> XMPP (the protocol) is also known as Jabber (the open federation). These
> two terms are used interchangeably. To learn more about Jabber/XMPP,
> see [joinjabber.org](https://joinjabber.org/).

> [!WARNING]
> **Create a dedicated user first. It will not relay messages from yourself if you use your account**

## Configuration

> [!TIP]
> For detailed information about xmpp settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[xmpp.myxmpp]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
Server="jabber.example.com:5222"
Jid="user@example.com"
Password="yourpass"
Muc="conference.jabber.example.com"
Nick="xmppbot"
```
