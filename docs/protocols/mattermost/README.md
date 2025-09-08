# Mattermost

- Status: ???
- Maintainers: ???
- Features: ???

> [!WARNING]
> **Create a dedicated user first. It will not relay messages from yourself if you use your account**

## Configuration

> [!TIP]
> For detailed information about mattermost settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[mattermost.mymattermost]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
# Don't use http or https scheme in the Server
Server="yourmattermostserver.domain:443"
# Uncomment if you're accessing your server over plaintext HTTP
#NoTLS=true

#the team name as can be seen in the mattermost webinterface URL
#in lowercase, without spaces
Team="yourteam"

Login="yourlogin"
Password="yourpass"

PrefixMessagesWithNick=true
PreserveThreading=true
```

## FAQ 

### "version not supported error"

By default, matterbridge tries HTTPS to connect to your Mattermost setup.  
If you're using a test setup for Mattermost, this will probably listen on HTTP and on port 8065.

Add ```NoTLS=true``` and use ```Server="yourmattermostserver.domain:8065"``` to your Mattermost configuration.   

### Mattermost doesn't show the IRC nicks

If you're running the webhooks version, this can be fixed by either:
* enabling "override usernames". See [mattermost documentation](http://docs.mattermost.com/developer/webhooks-incoming.html#enabling-incoming-webhooks)
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.toml.

If you're running the login/password version, you'll need to:
* setting ```PrefixMessagesWithNick``` to ```true``` in ```mattermost``` section of your matterbridge.toml.

Also look at the ```RemoteNickFormat``` setting.

### Session expire

Use personal access tokens [that don't expire](https://docs.mattermost.com/developer/personal-access-tokens.html), or [increase the session timeout](https://forum.mattermost.org/t/solved-removing-the-session-timeout/3033).
