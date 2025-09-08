# Matterbridge API settings

> [!TIP]
> This page contains the details about matterbridge API settings. More general information about the matterbridge API can be found in [README.md](README.md).

## BindAddress 

Address to listen on for API

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  BindAddress="127.0.0.1:4242"
  ```

## Buffer

Amount of messages to keep in memory

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *int*
- Default: *10*
- Example:
  ```toml
  Buffer=1000
  ```

## Token

HTTP Bearer token used for authentication. If unset, no authentication
will be applied at all and anyone who can reach the API will be able
to control your matterbridge instance.

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *string*
- Example:
  ```toml
  Token="mytoken"
  ```
