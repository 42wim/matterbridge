# RocketChta

- Status: ???
- Maintainers: ???
- Features: ???

> [!WARNING]
> **Create a dedicated user first. It will not relay messages from yourself if you use your account**

## Configuration

> [!TIP]
> For detailed information about rocketchat settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[rocketchat.myrocketchat]
# Server/Login/Password are only required when not using webhooks
Server="https://rocket.example.com:443"
# When using token authentication, Login is your ID not your email
Login="yourlogin@domain.com"
Password="yourpass"

# if you're using login/pass you can better enable because of this bug:
# https://github.com/RocketChat/Rocket.Chat/issues/7549
PrefixMessagesWithNick=true
```

## FAQ

### What is `You are not authorized to change message properties` error?

Your matterbridge bot must have the role `bot`. If it doesn't you will get the bug described in [#16097](https://github.com/RocketChat/Rocket.Chat/issues/16097).

