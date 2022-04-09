# VK SDK for Golang

[![PkgGoDev](https://pkg.go.dev/badge/github.com/SevereCloud/vksdk/v2/v2)](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2?tab=subdirectories)
[![VK Developers](https://img.shields.io/badge/developers-%234a76a8.svg?logo=VK&logoColor=white)](https://dev.vk.com/)
[![codecov](https://codecov.io/gh/SevereCloud/vksdk/branch/master/graph/badge.svg)](https://codecov.io/gh/SevereCloud/vksdk)
[![VK chat](https://img.shields.io/badge/VK%20chat-%234a76a8.svg?logo=VK&logoColor=white)](https://vk.me/join/AJQ1d6Or8Q00Y_CSOESfbqGt)
[![release](https://img.shields.io/github/v/tag/SevereCloud/vksdk?label=release)](https://github.com/SevereCloud/vksdk/releases)
[![license](https://img.shields.io/github/license/SevereCloud/vksdk.svg?maxAge=2592000)](https://github.com/SevereCloud/vksdk/blob/master/LICENSE)

**VK SDK for Golang** ready implementation of the main VK API functions for Go.

[Russian documentation](https://github.com/SevereCloud/vksdk/wiki)

## Features

Version API 5.131.

- [API](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/api)
  - 500+ methods
  - Ability to modify HTTP client
  - Request Limiter
  - Support [zstd](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/api#VK.EnableZstd)
and [MessagePack](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/api#VK.EnableMessagePack)
  - Token pool
  - [OAuth](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/api/oauth)
- [Callback API](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/callback)
  - Tracking tool for users activity in your VK communities
  - Supports all events
  - Auto setting callback
- [Bots Long Poll API](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/longpoll-bot)
  - Allows you to work with community events in real time
  - Supports all events
  - Ability to modify HTTP client
- [User Long Poll API](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/longpoll-user)
  - Allows you to work with user events in real time
  - Ability to modify HTTP client
- [Streaming API](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/streaming)
  - Receiving public data from VK by specified keywords
  - Ability to modify HTTP client
- [FOAF](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/foaf)
  - Machine-readable ontology describing persons
  - Works with users and groups
  - The only place to get page creation date
- [Games](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/games)
  - Checking launch parameters
  - Intermediate http handler
- [VK Mini Apps](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/vkapps)
  - Checking launch parameters
  - Intermediate http handler
- [Payments API](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/payments)
  - Processes payment notifications
- [Marusia Skills](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/marusia)
  - For creating Marusia Skills
  - Support SSML

## Install

```bash
# go mod init mymodulename
go get github.com/SevereCloud/vksdk/v2@latest
```

## Use by

- A simple chat bridge: <https://github.com/42wim/matterbridge>
- [Joe](https://github.com/go-joe/joe) adapter: <https://github.com/tdakkota/joe-vk-adapter>
- [Logrus](https://github.com/sirupsen/logrus) hook: <https://github.com/SevereCloud/vkrus>

### Example

```go
package main

import (
	"context"
	"log"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

func main() {
	token := "<TOKEN>" // use os.Getenv("TOKEN")
	vk := api.NewVK(token)

	// get information about the group
	group, err := vk.GroupsGetByID(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Initializing Long Poll
	lp, err := longpoll.NewLongPoll(vk, group[0].ID)
	if err != nil {
		log.Fatal(err)
	}

	// New message event
	lp.MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		if obj.Message.Text == "ping" {
			b := params.NewMessagesSendBuilder()
			b.Message("pong")
			b.RandomID(0)
			b.PeerID(obj.Message.PeerID)

			_, err := vk.MessagesSend(b.Params)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	// Run Bots Long Poll
	log.Println("Start Long Poll")
	if err := lp.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## LICENSE

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FSevereCloud%2Fvksdk.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FSevereCloud%2Fvksdk?ref=badge_large)
