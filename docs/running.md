# Running matterbridge

This page is about running matterbridge from CLI, or as a service. For setup, see [setup.md](setup.md). For configuration, see [config.md](config.md).

## Command-line

```bash
Usage of ./matterbridge:
  -conf string
        config file (default "matterbridge.toml")
  -debug
        enable debug
  -gops
        enable gops agent
  -version
        show version
```

## docker-compose image

From the directory where you have your configuration `matterbridge.toml`, create a file named `docker-compose.yml`:

```yml
version: '3.7'
services:
  matterbridge:
    image: 42wim/matterbridge:stable
    restart: unless-stopped
    volumes:
    - ./matterbridge.toml:/etc/matterbridge/matterbridge.toml:ro
#    command: -debug
```

Afterwards, start the container with `docker-compose up -d`.

# Service integrations

## systemd

A sample systemd unit to run matterbridge in the background, restarting if necessary.

`/etc/systemd/system/matterbridge.service`, remember to correct `ExecStart=` and `User=`

```dosini
[Unit]
Description=Matterbridge daemon
After=network-online.target

[Service]
Type=simple
ExecStart=/path/to/your/matterbridge/binary -conf /path/to/your/matterbridge.toml
Restart=always
RestartSec=5s
User=matterbridge

[Install]
WantedBy=multi-user.target
```

Reload systemd with `sudo systemctl daemon-reload`, enable on system start and now with `sudo systemctl enable --now matterbridge.service`

## OpenRC

A sample OpenRC service to run matterbridge in the background.

/etc/init.d/matterbridge
```sh
#!/sbin/openrc-run
# Copyright 2021-2022 Gentoo Authors
# Distributed under the terms of the GNU General Public License v2

command=/usr/bin/matterbridge
command_args="-conf ${MATTERBRIDGE_CONF:-/etc/matterbridge/matterbridge.toml} ${MATTERBRIDGE_ARGS}"
command_user="mattermost:mattermost"
pidfile="/run/${RC_SVCNAME}.pid"
command_background=1

depend() {
	need net
}
```
Enable with `sudo rc-update add matterbridge default`

Start with `sudo rc-service matterbridge start`
