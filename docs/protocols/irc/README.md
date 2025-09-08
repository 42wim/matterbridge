# IRC

- Status: Working
- Maintainers: @poVoq, @selfhoster1312
- Features: ???

## Configuration

> [!TIP]
> For detailed information about irc settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[irc.myirc]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
Server="irc.libera.chat:6667"
Nick="yourbotname"
Password="yourpassword"
# Enable SASL on modern servers like irc.libera.chat
# UseSASL=true
```

## FAQ

### How to connect to a password-protected channel?

```toml
[[gateway.inout]]
account="irc.myirc"
channel="#some-passworded-channel"
options = { key="password" }
```

### How to connect to OFTC-style NickServ

```toml
[irc.myirc]
Nick="yournick"
Server="irc.oftc.net:6697"
RunCommands=["PRIVMSG nickserv :IDENTIFY yourpass yournick"]
```