# Keybase

- Status: ???
- Maintainers: ???
- Features: ???

> [!WARNING]
> **Remember to create a dedicated Keybase user (a new account specifically for the bot).**
```

## Configuration

**Basic configuration example:**

```toml
[keybase.mykeybase]
# Your bot account MUST already have access to the provided team or subteam!
Team="mykeybase.team"
RemoteNickFormat="[{PROTOCOL}/{BRIDGE}] <{NICK}> "
```

## FAQ

### How to create a new account for the bot?

Install Keybase on the server running matterbridge (https://keybase.io/download)
Log into the bot account on your server

## Headless server install + systemd

If you run matterbridge as a service in systemd and want to use a bridge with keybase, be sure to run the service as the default user via *user.slice*, since keybase also runs its services in *user.slice*.

Create a `matterbridge-user.service` in `~/.config/systemd/user/`:

```
[Unit]
Description=matterbridge
After=network.target

[Service]
ExecStart=/path-to-your-executable/matterbridge
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

Then add the `--user` argument to all your `systemctl` commands. For example, `systemctl --user enable --now matterbridge-user.service`.

> [!WARNING]
> **If you are using an SSH session, user services might shut down, use this to prevent:**  `loginctl enable-linger`