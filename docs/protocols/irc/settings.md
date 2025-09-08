# IRC settings

> [!TIP]
> This page contains the details about IRC settings. More general information about IRC support in matterbridge can be found in [README.md](README.md).

## Charset

If you know your charset, you can specify it manually. Otherwise it tries to detect this automatically.
 
The selected charset will be converted to utf-8 when sent to other bridges.

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  Charset="utf-8"
  ```

## ColorNicks

ColorNicks will show each nickname in a different color.
Only works in IRC right now.

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  ColorNicks=true
  ```

## DebugLevel

Debug log verbosity.

- Setting: **OPTIONAL**
- Default: *0*
- Format: *int*
- Example:
  ```toml
  DebugLevel=1
  ```

## JoinDelay

Delay in milliseconds between channel joins.
Only useful when you have a LOT of channels to join.

- Setting: **OPTIONAL**, **RELOADABLE**
- Default: *0*
- Format: *int*
- Example:
  ```toml
  JoinDelay=1000
  ```

## MessageDelay

Flood control.
Delay in milliseconds between each message send to the IRC server.

- Setting: **OPTIONAL**, **RELOADABLE**
- Default: *1300*
- Format: *int*
- Example:
  ```toml
  MessageDelay=1300
  ```

## MessageLength

Maximum length of message sent to irc server. If it exceeds
`<message clipped>` will be add to the message.

- Setting: **OPTIONAL**, **RELOADABLE**
- Default: *400*
- Format: *int*
- Example:
  ```toml
  MessageLength=400
  ```

## MessageQueue

Maximum amount of messages to hold in queue. If queue is full 
messages will be dropped. 
`<message clipped>` will be add to the message that fills the queue.

- Setting: **OPTIONAL**, **RELOADABLE**
- Default: *30*
- Format: *int*
- Example:
  ```toml
  MessageQueue=30
  ```

## MessageSplit
Split messages on `MessageLength` instead of showing the `<message clipped>`
WARNING: this could lead to flooding

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *boolean*
- Example:
  ```toml
  MessageSplit=true
  ```

## Nick *
Your nick on irc. 

- Setting: REQUIRED
- Format: *string*
- Example:
  ```toml
  Nick="matterbot"
  ```

## NickServNick
If you registered your bot with a service like Nickserv on freenode.
Also being used when `UseSASL=true`
Note: when `UseSASL=true`, this is the name of *your* account.
Note: if you want do to quakenet auth, set NickServNick="Q@CServe.quakenet.org"

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  NickServNick="nickserv"
  ```

## NickServPassword
The password you use if you registered your bot with a service like Nickserv on freenode.
Also being used when `UseSASL=true`
Also see `NickServNick`

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  NickServPassword="secret"
  ```

## NickServUsername
Only used for quakenet auth.
See https://github.com/42wim/matterbridge/issues/263 for more info

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  NickServUsername="username"
  ```

## NoSendJoinPart
Do not send joins/parts to other bridges
Currently works for messages from the following bridges: irc, mattermost, slack

- Setting: **OPTIONAL**
- Format: *boolean*
- Example:
  ```toml
  NoSendJoinPart=true
  ```

## Password 

Password for irc server (if necessary)

- Setting: **OPTIONAL**
- Format: *string*
- Example:
  ```toml
  Password="s3cret"
  ```

## Pingdelay

PingDelay specifies how long to wait to send a ping to the irc server.
You can use s for second, m for minute

- Setting: **OPTIONAL**, **RELOADABLE**
Default: "1m"
- Example:
  ```toml
  PingDelay="1m"
  ```

## RejoinDelay
Delay in seconds to rejoin a channel when kicked

- Setting: **OPTIONAL**, **RELOADABLE**
Default: 0
- Format: *int*
- Example:
  ```toml
  RejoinDelay=2
  ```

## RunCommands

RunCommands allows you to send RAW irc commands after connection

- Setting: **OPTIONAL**, **RELOADABLE**
- Format: *List<string>*
- Example:
  ```toml
  RunCommands=["PRIVMSG user :hello","PRIVMSG chanserv :something"]
  ```

## Server

irc server to connect to. 

- Setting: REQUIRED
- Format: *string* (hostname:port)
- Example:
  ```toml
  Server="irc.freenode.net:6667"
  ```

## StripMarkdown

Strips Markdown from messages

- Setting: **OPTIONAL**
- Format: *boolean*
- Example:
  ```toml
  StripMarkdown=true
  ```

## UseSASL

Enable SASL (PLAIN) authentication. (freenode requires this from eg AWS hosts)
It uses `NickServNick` and `NickServPassword` as login and password

- Setting: **OPTIONAL**
- Format: *boolean*
- Example:
  ```toml
  UseSASL=true
  ```

## UseTLS

Enable to use TLS connection to your irc server. 

- Setting: **OPTIONAL**
- Format: *boolean*
- Example:
  ```toml
  UseTLS=true
  ```

## UseRelayMsg

Enable to replace bot's nick with user's nick.
- `RemoteNickFormat` has to contain `/`.
The server has to support RELAYMSG.
Bot may need to be channel operator to use RELAYMSG.

- Setting: OPTIIONAL
- Format: *boolean*
- Example:
  ```toml
  UseRelayMsg=true
  ```

## VerboseJoinPart

Enable to show verbose users joins/parts (ident@host) from other bridges
Currently works for messages from the following bridges: irc

- Setting: **OPTIONAL **
- Format: *boolean*
- Example:
  ```toml
  VerboseJoinPart=true
  ```
