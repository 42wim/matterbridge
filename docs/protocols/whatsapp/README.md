# Whatsapp

- Status: ???
- Maintainers: ???
- Features: ???

## Configuration

> [!TIP]
> For detailed information about whatsapp settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[whatsapp.mywhatsapp]
RemoteNickFormat="[{PROTOCOL}] @{NICK}: "
# Get a disposable SIM card
Number="+48111222333"
# See FAQ
SessionFile="session-48111222333.gob"
# If your terminal uses a light color scheme, uncommebt below
#QrOnWhiteTerminal=true
```

## FAQ

### What is the QR-code about?

The QR-Code is created by matterbridge on the terminal: It's the way WhatsApp authenticates any WA-Web session (which the bridge is using). At startup, matterbridge will prompt a QR-code for about 1 minute that you have to scan from your WA-instance (on your phone or android-VM). To scan it, go to Settings > WhatsApp Web in your WA client (if you run a VM, you may need to connect your webcam, or make a virtual one, see link above):

![](https://user-images.githubusercontent.com/8722846/64077252-ae3be180-ccce-11e9-812e-6e8a1e833346.png)

### How to set the WA-channel in the configuration file of matterbridge? 

To setup a gateway between two protocols in matterbridge, you need to specify a channel for WA that should be bridged. The chat/group titles you see in WA won't work (e.g. from the screenshot above, "Test" won't work). Fortunately, matterbridge will complain about improper channel names and suggest the correct ones to you. In my case, it was a string of mainly numbers including the phone number of who created the group chat. Maybe there is a way to get this out of WA?
(Currently, the WA example config does not explain this at all; it doesn't even mention the gateway part.)

### How to set a nice channel name?

Use `tengo`:

Save below snippet as `myremotenickformat.tengo`
```
if channel == "48111222333-1549986983@g.us" {
        result = "[CfP #Code for Pakistan] @"+nick
}
```
Add `RemoteNickFormat="{TENGO}"` to the specific bridge

and add the following in matterbridge.toml

```
[tengo]
RemoteNickFormat="myremotenickformat.tengo"
```

For context see: https://github.com/42wim/matterbridge/issues/725
