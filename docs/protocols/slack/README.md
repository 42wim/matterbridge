# Slack

- Status: ???
- Maintainers: ???
- Features: ???

## Configuration

**Basic configuration example:**

```toml
[slack]
[slack.myslack]
RemoteNickFormat="{BRIDGE} - @{NICK}"
Token="#####"
# this will maps threads from other bridges on slack threads
PreserveThreading=true
```

## FAQ

### How to get create an account for my matterbridge bot?

See [account.md](account.md).

## Auth Comparison

This is a simple comparison of the Slack connection methods, to help you understand and differentiate between each:

| 👇 Feature \ Token Type 👉 | Bot user | Legacy<br/>(personal account) | Legacy<br/>(dedicated account) |
|---|:---:|:---:|:---:|
| Bridge Type | `slack` | `slack-legacy` | `slack-legacy` |
| Token prefix | `xoxb-` | `xoxp-` | `xoxp-` |
| Supported | ✅ |   | ✅ |
| 1. Auto-join channels |   | ✅ | ✅ |
| 3. Uses separate integration slot | :x: |   |   |
| 4. User's outgoing messages relayed | ✅ |   | ✅ |
| 5. File uploads show as from... | bot | you | dedicated user |# Slack

## Messages come from Slack API tester
Did you set `RemoteNickFormat`?    
Try adding `RemoteNickFormat="<{NICK}>"`

## Messages from other bots aren't getting relayed

If you're using `WebhookURL` in your Slack configuration, this is normal.
If you only have `Token` configuration, this could be a bug. Please open an issue.
