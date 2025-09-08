# Matrix

- Status: ???
- Maintainers: @poVoq
- Features: ???

> [!WARNING]
> **Create a dedicated user first. It will not relay messages from yourself if you use your account**

## Configuration

> [!TIP]
> For detailed information about matrix settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[matrix.mymatrix]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
Server="https://matrix.org"
Login="yourlogin"
Password="yourpass"
# Alternatively, you can use MXID + a session Token
#MxID="@yourbot:example.net"
#Token="tokenforthebotuser"
```

## FAQ

### How to encrypt matterbridge messages to Matrix?

Matterbridge doesn't properly encrypt its messages. So although matterbridge *does* work with matrix, even with matrix' unencrypted rooms, the messages sent by matterbridge will all show a warning symbol to everyone, something about "WARNING: This message was sent unencrypted!", which might irritate users.

So there is a need for something that sits in the middle, pretends to be a matrix server (so that matterbridge can talk to it), and can forward everything to the real matrix server (so that the messages actually arrive), and also magically transparently "encrypts" everything (so that the messages show no "unencrypted" warning). This is exactly what pantalaimon does. Keep in mind that this effectively means you do a MITM-attack on yourself, so the connection between matterbridge and pantalaimon is *basically plaintext* and very vulnerable. You really should run matterbridge and pantalaimon on the same machine, and make sure that pantalaimon is only accessible to yourself. (I don't know if VPS is a problem here, so if you are running on a VPS then think twice before you do this setup.)

#### bridge.toml

```toml
[general]
MediaDownloadPath="/path/to/http/server/"
MediaServerDownload="https://foo.bar.org/server/"
MediaDownloadSize=10000000

[telegram.mytelegram]
Token="1234567890:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
RemoteNickFormat="{NICK}@{PROTOCOL}: "
MediaConvertWebPToPNG=true
MediaConvertTgs="png"
#QuoteFormat="{MESSAGE} (re @{QUOTENICK}: {QUOTEMESSAGE})"
UseFirstName=true

[matrix.mymatrix]
# Server="https://matrix.org"
# Matterbridge does not support encrypted group chats.
# Therefore, use Pantalaimon to MiTM myself:
Server="http://localhost:20662"
# Dedicated user
# Messages sent from this user will not be relayed to avoid loops.
Login="mybot"
Password="abcdefghijklmnopqrstuvwxyz"
RemoteNickFormat="{NICK}@{PROTOCOL}: "
#Whether to send the homeserver suffix. eg ":matrix.org" in @username:matrix.org
#to other bridges, or only send "username".(true only sends username)
NoHomeServerSuffix=true
HTMLDisable=true

[[gateway]]
name="foobar"
enable=true

[[gateway.inout]]
account="telegram.mytelegram"
channel="-1234567890123"

[[gateway.inout]]
account="matrix.mymatrix"
channel="!abcdefghijklmnopqr:matrix.org"
```

#### pantalaimon.conf

```ini
[Default]
LogLevel = Debug
SSL = True

[local-matrix]
Homeserver = https://matrix.org
ListenAddress = localhost
ListenPort = 20662
SSL = False
IgnoreVerification = True
UseKeyring = False
```

#### run_pantalaimon.sh

In theory, it suffices to just call `dbus-run-session -- pantalaimon --config pantalaimon.conf`

However, I want *all* the logs, so I run this:

```sh
dbus-run-session -- pantalaimon --log-level debug --config pantalaimon.conf 2>&1 | \
    ./tee_unless_regex.py 'INFO: pantalaimon: Trying to decrypt sync|INFO: pantalaimon: Decrypting sync' \
    2> pantalaimon_$(date +%s).log
```

#### tee_unless_regex.py

```
#!/usr/bin/env python3

import re
import sys

def run_regex(regex):
	while True:
		try:
			line = sys.stdin.readline()
		except KeyboardInterrupt:
			# Ctrl-C
			return
		if not line:
			# EOF
			return
		line = line.rstrip('\n')
		print(line)
		if not regex.search(line):
			print(line, file=sys.stderr)
			sys.stderr.flush()  # This flush() is the entire reason why I don't just use 'grep -v'. Somehow, unbuffer+grep just doesn't work. But why!?


def run():
	if len(sys.argv) != 2:
		print('USAGE: {} <SOME_REGEX>'.format(argv[0]))
		exit(1)

	run_regex(re.compile(sys.argv[1]))


if __name__ == '__main__':
	run()
```

#### Setup

There are setup-steps missing. In particular, you absolutely need pactl at some point. TODO: Please fill in these details.

#### Invocation

In one screen: `./run_pantalaimon.sh`

In another screen: `./matterbridge-THEVERSION-linux-arm -conf bridge.toml -debug | tee bridge_$(date +%s).log`

(Again, the `-debug | …` stuff isn't necessary, but I personally want permanent logs of everything, just so I can trace back if something ever goes wrong. And I suggest that you do that, too.)