# SSHChat

- Status: ???
- Maintainers: ???
- Features: ???

> [!WARNING]
> In the gateway setup you should set channel to "sshchat" in order for things to work properly.

**Basic configuration example:**

```toml
[sshchat.mychat]
Server="address:port" # eg "localhost:2022" or "1.2.3.4:22"
Nick="matterbridge"
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
```
