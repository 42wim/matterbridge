# MSTeams

- Status: ???
- Maintainers: ???
- Features:
  - edits: no
  - attachments: no

## Configuration

**Basic configuration example:**

```toml
[msteams.teams]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
TenantID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
ClientID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
TeamID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

## FAQ

### How to get the TenantID, ClientID, and TeamID?

See [account.md](account.md).

### How to send HTML/codeblock/emoji to MSTeams?

That's not possible due to Microsoft API limitations.

They will not be rendered properly, sorry.

### Why is feature X not supported?

Not supported by Microsoft API, sorry.

### How to receive messages with attachments and images

Not supported by Microsoft API (TODO: is that still the case?)
