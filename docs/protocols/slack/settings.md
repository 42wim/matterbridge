# Slack settings

> [!TIP]
> This page contains the details about slack settings. More general information about slack support in matterbridge can be found in [README.md](README.md).

## Debug

Extra slack specific debug info, warning this generates a lot of output.

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  Debug=true
  ```

## PreserveThreading

Opportunistically preserve threaded replies between Slack channels.
This only works if the parent message is still in the cache.
Cache is flushed between restarts.
Note: Not currently working on gateways with mixed bridges of
both slack and slack-legacy type. Context in issue #624.

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  PreserveThreading=true
  ```

## ShowUserTyping

Enable showing "user_typing" events from across gateway when available.
Protip: Set your bot/user's "Full Name" to be "Someone (over chat bridge)",
and so the message will say "Someone (over chat bridge) is typing".

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  ShowUserTyping=true
  ```

## SyncTopic

Enable to sync topic/purpose changes from other bridges
Only works syncing topic changes from slack bridge for now

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  SyncTopic=true
  ```

## Token

Token to connect with the Slack API

- Setting: **REQUIRED** (when not using webhooks)
- Format: *string*
- Example:
  ```toml
  Token="yourslacktoken"
  ```

### IconURL

> [!WARNING]
> NOT RECOMMENDED TO USE INCOMING/OUTGOING WEBHOOK.
> USE DEDICATED BOT USER WHEN POSSIBLE!

Icon that will be showed in slack.
The string "{NICK}" (case sensitive) will be replaced by the actual nick / username.
The string "{BRIDGE}" (case sensitive) will be replaced by the sending bridge.
The string "{LABEL}" (case sensitive) will be replaced by label= field of the sending bridge.
The string "{PROTOCOL}" (case sensitive) will be replaced by the protocol used by the bridge.

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  IconURL="https://robohash.org/{NICK}.png?size=48x48"
  ```

### WebhookBindAddress

> [!WARNING]
> NOT RECOMMENDED TO USE INCOMING/OUTGOING WEBHOOK.
> USE DEDICATED BOT USER WHEN POSSIBLE!

Address to listen on for outgoing webhook requests from slack.
See account settings - integrations - outgoing webhooks on slack

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  WebhookBindAddress="0.0.0.0:9999"
  ```

### WebhookURL

> [!WARNING]
> NOT RECOMMENDED TO USE INCOMING/OUTGOING WEBHOOK.
> USE DEDICATED BOT USER WHEN POSSIBLE!

Url is your incoming webhook url as specified in slack.
See account settings - integrations - incoming webhooks on slack

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  WebhookURL="https://hooks.slack.com/services/yourhook"
  ```
