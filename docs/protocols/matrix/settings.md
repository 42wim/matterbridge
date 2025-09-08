# Matrix settings

> [!TIP]
> This page contains the details about matrix settings. More general information about matrix support in matterbridge can be found in [README.md](README.md).

## HTMLDisable

Whether to disable sending of HTML content to matrix
See https://github.com/42wim/matterbridge/issues/1022

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  HTMLDisable=true
  ```

## Login

login of your bot.
Use a dedicated user for this and not your own!
Messages sent from this user will not be relayed to avoid loops.

- Setting: **REQUIRED**
- Format: *string*
- Example: 
  ```toml
  Login="yourlogin"
  ```

## NoHomeServerSuffix

Whether to send the homeserver suffix. eg ":matrix.org" in @username:matrix.org
to other bridges, or only send "username".(true only sends username)

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  NoHomeServerSuffix=true
  ```

## Password

password of your bot.

- Setting: **REQUIRED**
- Format: *string*
- Example: 
  ```toml
  Password="yourpass"
  ```

## Server

Server is your homeserver (eg https://matrix.org)

- Setting: **REQUIRED**
- Format: *string*
- Example: 
  ```toml
  Server="https://matrix.org"
  ```

## UseUserName

Shows the username instead of the displayname

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  UseUserName=true
  ```
