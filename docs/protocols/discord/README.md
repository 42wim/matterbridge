# Discord

- Status: ???
- Maintainers: ???
- Features: ???

## Configuration

> [!TIP]
> For detailed information about discord settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[discord]
[discord.mydiscord]
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
Token="######"
Server="name or uid of guild"
AutoWebhooks=true
# Map threads from other bridges on discord replies
PreserveThreading=true
```

## FAQ

### How to create a discord bot token for matterbridge?

See [account.md](account.md).

### Username/avatar spoofing

[Creating a message](https://discordapp.com/developers/docs/resources/channel#create-message) via a user's API token (the basic configuration above) only lets Matterbridge post with the user/avatar that generated the token. But [executing a webhook](https://discordapp.com/developers/docs/resources/webhook#execute-webhook) can set any username and avatar URL.

If you grant the bot the "Manage Webhooks" permission, it will automatically load and create webhooks in every bridged channel. You can even grant that permission on specific channels, if you don't want to give it global permission.

1. Server Settings -> Roles
2. Select your app's role (see [here](https://support.discord.com/hc/en-us/articles/206029707-How-do-I-set-up-Permissions-) for more info about how to create roles)
3. Grant "Manage Webhooks"

Then, just provide `AutoWebhooks=true` for the Discord account settings, as shown below:

```toml
[discord]
    [discord.mydiscord]
    Token="######"
    Server="Wumpus Technologies Inc."
    AutoWebhooks=true
```


<details><summary>If you would like to create webhooks yourself, instead of letting the bot manage them, follow these instructions.</summary>

1. On Discord, go to Server Settings, then Integrations, then Webhooks -> "Create Webhook"
2. Specify the name and channel, and copy the resulting webhook URL
3. Paste in your gateway as `WebhookURL="https://discordapp.com/api/webhooks/529689699999999999/Da-H4RRY_P0-kjdsknkfgfjghf`

Specify a webhook per channel:

```toml
[[gateway]]
name="testing"
enable=true

[[gateway.inout]]
account="slack.myworkspace"
channel="testing"

[[gateway.inout]]
account="discord.myserver"
channel="testing"
    
    # Specify options for this gateway link
    [gateway.inout.options]
    WebhookURL="https://discordapp.com/api/webhooks/thing1/thing2"
```

</details>

### Guessing avatars when they are missing

> **This feature is only available when sending messages using webhooks.**

Avatars from source platforms will usually be shown in Discord messages. However, sometimes users can't (or don't) set their avatar on the source platform, (e.g for messages sent from IRC to Discord), so messages from those may use a default avatar.

The `UseLocalAvatar` specifies source bridges for which an avatar should be "guessed", so that all messages have avatars. This works by comparing the source message username to an existing Discord user on your server, and using the avatar of that Discord user.

Note that it won't try to "guess" avatars when:

- an avatar on the source platform is present, or
- there are multiple Discord users with the same name.

As shown below, you can either provide the bridge platform name (`"irc"`) or the full account name (`"slack.myworkspace"`). This means that for messages coming from IRC, or messages from the Slack `myworkspace` account, avatar "guessing" is enabled.

```toml
[discord]
    [discord.mydiscord]
    Token="#####"
    Server="Wumpus Technologies Inc."
    UseLocalAvatar=["irc", "slack.myworkspace"]
```

### Seeing IDs instead of usernames on the other bridges?

If you want roles/groups mentions to be shown with names instead of ID,  you'll need to give your bot the "Manage Roles" permission.

### Can I join the bot to a server I don't own?

You need permissions on the server to be able to join the bot, if you're allowed to here's how you do it:
Click on the bot here: https://discordapp.com/developers/applications/me

Click the "Generate OAUTH" button, copy the URL and paste it into your address bar, open the page and you'll see the drop down menu for adding the bot to any server you have permissions on.

### Messages not being edited, but a new message with (edit) appears?
Message editing is not supported by discord using webhooks. If you want message editing, disable webhooks and use token only

### Error obtaining server members: HTTP 403 Forbidden

Discord changed its API:

See <https://github.com/42wim/matterbridge/wiki/Discord-bot-setup#privileged-gateway-intents> for a fix  
See <https://github.com/42wim/matterbridge/issues/1263> for more info

### Do I need to allow inbound connections for webhooks to work

No. 

### Replies show the botname instead of the username/avatar

This is a discord limitation: https://github.com/42wim/matterbridge/issues/1558#issuecomment-1030713994
You can do a workaround by setting `PreserveThreading=false`
