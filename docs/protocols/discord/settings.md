# Discord settings

> [!TIP]
> This page contains the details about discord settings. More general information about discord support in matterbridge can be found in [README.md](README.md).

## Server

Name or uid of your server/guild.

For example, If you want to bridge `https://discord.com/channels/AAAA/BBBB`, then this should be `AAAA`.

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Server="yourservername"
  ```

## Token

Token to connect with Discord API. See [account.md](account.md) for
instructions to generate your token. If you want roles/groups mentions to be
shown with names instead of ID, you'll need to give your bot the "Manage Roles"
permission.

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Token="YOUR_TOKEN_HERE"
  ```

## AllowMention

AllowMention controls which mentions are allowed.

If not specified, all mentions are allowed. Note that even when a mention is
not allowed, it will still be displayed nicely and be clickable. It just
prevents the ping/notification.

- Setting: **OPTIONAL**
- Format: *List[string]*
- Possible values:
  - `"everyone"` allows `@everyone` and `@here` mentions
  - `"roles"` allows `@role` mentions
  - `"users"` allows `@user` mentions
- Example:
  ```toml
  AllowMention=["everyone", "roles", "users"]
  ```

## ShowEmbeds

Shows title, description and URL of embedded messages (sent by other bots)

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  ShowEmbeds=true
  ```

## UseUserName

Shows the username instead of the server nickname

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  UseUserName=true
  ```

## UseDiscriminator

Show `#xxxx` discriminator with `UseUserName`

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  UseDiscriminator=true
  ```

## WebhookURL

Specify WebhookURL. If given, will relay messages using the Webhook, which
gives a better look to messages.

If you have multiple discord channels, it is recommended to use the
`AutoWebHooks` setting. Alternatively, you'll have to specify the webhook URL
for each gateway and not in the account configuration.

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  `WebhookURL="Yourwebhooktokenhere"` 

## AutoWebhooks

Relay messages using the Webhook, which gives a better look to messages. Needs "Manage Webhooks" permission to function.

AutoWebhooks automatically configures message sending in the style of puppets.
This is an easier alternative to manually configuring "WebhookURL" for each gateway,
as turning this on will automatically load or create webhooks for each channel.
This feature requires the "Manage Webhooks" permission (either globally or as per-channel).


Setting: OPTIONAL \
Format: boolean \
Example:

`AutoWebhooks=true`
