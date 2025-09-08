# RocketChat settings

> [!TIP]
> This page contains the details about rocketchat settings. More general information about rocketchat support in matterbridge can be found in [README.md](README.md).

## Login

login needs to be the login with email address! user@domain.com
Use a dedicated user for this and not your own!

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Login="yourlogin@domain.com"
  ```

## Password

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Password="yourpass"
  ```

## Server 

The rocketchat hostname. (prefix it with http or https)

- Setting: **REQUIRED**
- Format: *string*
- Example:
  ```toml
  Server="https://yourrocketchatserver.domain.com:443"
  ```

### WebhookBindAddress

Address to listen on for outgoing webhook requests from rocketchat.
See administration - integrations - new integration - outgoing webhook

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  WebhookBindAddress="0.0.0.0:9999"
  ```

### Nick

Your nick/username as specified in your incoming webhook "Post as" setting

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  Nick="matterbot"
  ```

### WebhookURL

Url is your incoming webhook url as specified in rocketchat ([docs](https://rocket.chat/docs/administrator-guides/integrations/#how-to-create-a-new-incoming-webhook)).
See administration - integrations - new integration - incoming webhook

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  WebhookURL="https://yourdomain/hooks/yourhookkey"
  ```
