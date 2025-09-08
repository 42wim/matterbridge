# Zulip settings

> [!TIP]
> This page contains the details about zulip settings. More general information about zulip support in matterbridge can be found in [README.md](README.md).

## Login

Username of the bot, normally called yourbot-bot@yourserver.zulipchat.com.

See `username` in `Settings` → `Your bots`.

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Login="yourbot-bot@yourserver.zulipchat.com"
  ```

## Server

URL of your zulip instance.

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Server="https://yourserver.zulipchat.com"
  ```

## Token

Zulip API token (called bot API key in `Settings` → `Your bots`).

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Token="Yourtokenhere"
  ```
