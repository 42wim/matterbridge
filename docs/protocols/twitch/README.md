# Twitch

- Status: ???
- Maintainers: ???
- Features: ???

## Configuration

> [!WARNING]
> Twitch gets A LOT OF messages in almost any chatroom and this will be an
> issue for other bridges that will ratelimit/ban you.   
> So be sure to do this only in rooms with not much users

> [!WARNING]
> To relay messages to Twitch, The Twitch user will require mod powers
> for the channel/streamer specified in the gateway settings.

**Basic configuration example:**

```toml
[irc.twitch]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
Nick="yournick"
Password="oauth:yourtoken"
Server="irc.chat.twitch.tv:6697"
UseTLS=true
```

## FAQ

### How to get a Twitch OAuth token for my matterbridge bot?

> [!WARNING]
> The service we recommend strongly recommends against using it for more than
> development/testing purposes.

[Twitch Token Generator](https://twitchtokengenerator.com/) can generate OAuth tokens on demand for your bots.
