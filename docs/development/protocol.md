# Implementing a new protocol

This guide explains how to create a new protocol backend to support a new gateway/bridge in matterbridge.

## Step-by step list

- [ ] Create a new catalog in [`/bridge` folder](https://github.com/42wim/matterbridge/tree/master/bridge) and a main file named after the bridge you are creating, such as `whatsapp.go`
- [ ] Implement a [`Bridger` interface](https://github.com/42wim/matterbridge/blob/2cfd880cdb0df29771bf8f31df8d990ab897889d/bridge/bridge.go#L11-L16)
- [ ] [`gitter`](https://github.com/42wim/matterbridge/blob/master/bridge/gitter/gitter.go) is a relatively simple bridge that you might use as a reference to adapt
- [ ] Mention your bridge exists in [`/gateway/bridgemap/bridgemap.go`](https://github.com/42wim/matterbridge/blob/master/gateway/bridgemap/bridgemap.go)
- [ ] Divide functionality in several files, as it is done for [slack](https://github.com/42wim/matterbridge/tree/master/bridge)
  - `yourbridge.go` with main struct and implementation of the `Bridger` interface
  - `handlers.go` with handling messages incoming to Bridge
  - `helpers.go` for all the misc functions and helpers
- [ ] Minimal set of features is sending and receiving text messages working.
- [ ] Documentation
  - [ ] Add a [sample configuration](https://github.com/42wim/matterbridge/commit/6372d599b1ca2497aa49142d10496f345041b678#diff-0fcc5f77f08a4f4106d2da34c4dcd133) of your bridge to `matterbridge.toml.sample` and explain all the custom options
  - [ ] Add your bridge to README
  - [ ] Document all exported functions
- [ ] Run `golint` and `goimports` and clean the code
- [ ] Run `go mod vendor` to pull in all the vendor code
- [ ] Send a PR

## Features

Below is a feature list that you might copy to your issue.

Features:
- [ ] Connect to external service
- [ ] Get all active chats
- [ ] Check if chosen channels exist externally
- [ ] Connect to chosen channel
- [ ] Show nicknames in external service
- [ ] Show nicknames in relayed messages
- [ ] Test if multiple channels are working
- [ ] Show profile pictures from your bridge in relayed messages
- [ ] Show profile picture in your bridge
- [ ] Handle reply/thread messages
- [ ] Handle deletes
- [ ] Handle edits
- [ ] Handle notifications 
- [ ] Create a channel if it doesn't exist
- [ ] Sync channel metadata (name, topic, etc.)
- [ ] Document settings in `matterbridge.toml.sample`
- [ ] Document bridge in README
- [ ] Explain setting up the bridge process for users in the wiki
- [ ] Add screenshots from your bridge in the wiki
- [ ] Document code

Handle messages
- [ ] text from the bridge
- [ ] text to the bridge
- [ ] image
- [ ] audio
- [ ] video
- [ ] contacts?
- [ ] any other?


## FAQ

**How can I set the default RemoteNickFormat for a protocol so users don't have to do it in a config file?**

@42wim?

**Why on Slack I see bot name instead of remote username?**

Check if you:
- [ ] did set `Message.Username` on the message being relayed
- [ ] did set `RemoteNickFormat` in config file

**Sending message to the bridge don't work**

- [ ] Channels must match. While sending the message to the bridge make sure that you set the `config.Message.Channel` field to channel as it is mentioned in the config file.