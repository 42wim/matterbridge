# Zulip

- Status: ???
- Maintainers: ???
- Features:
  - attachments: no

## Configuration

> [!TIP]
> For detailed information about zulip settings, see [settings.md](settings.md)

> [!WARNING]
> Don't forget to make sure that the streams you want to bridge exit and
> are subscribed to by the new bot. Otherwise, things will break. You can
> subscribe the bot using the stream's settings, under "Stream membership".

**Basic configuration example:**

```toml
[zulip.myzulip]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
Token="Yourtokenhere"
#Login is the Username of the bot you've got when creating the bot
Login="yourbot-bot@yourserver.com"
#The URL of your zulip server
Server="https://yourserver.com"
```

## FAQ

### How to generate the zulip token for my matterbridge bot?

1. Navigate to Settings () -> Your bots -> Add a new bot. 
2. Select Generic for bot type, fill out the form and click on Create bot.
3. Token is the API key, you've got when creating the bot

![](https://snipboard.io/CvlUEH.jpg)
