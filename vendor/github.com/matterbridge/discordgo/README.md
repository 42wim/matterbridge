# DiscordGo 

[![GoDoc](https://godoc.org/github.com/bwmarrin/discordgo?status.svg)](https://godoc.org/github.com/bwmarrin/discordgo) [![Go report](http://goreportcard.com/badge/bwmarrin/discordgo)](http://goreportcard.com/report/bwmarrin/discordgo) [![Build Status](https://travis-ci.org/bwmarrin/discordgo.svg?branch=master)](https://travis-ci.org/bwmarrin/discordgo) [![Discord Gophers](https://img.shields.io/badge/Discord%20Gophers-%23discordgo-blue.svg)](https://discord.gg/0f1SbxBZjYoCtNPP) [![Discord API](https://img.shields.io/badge/Discord%20API-%23go_discordgo-blue.svg)](https://discord.gg/0SBTUU1wZTWT6sqd)

<img align="right" src="http://bwmarrin.github.io/discordgo/img/discordgo.png">

DiscordGo is a [Go](https://golang.org/) package that provides low level 
bindings to the [Discord](https://discordapp.com/) chat client API. DiscordGo 
has nearly complete support for all of the Discord API endpoints, websocket
interface, and voice interface.

If you would like to help the DiscordGo package please use 
[this link](https://discordapp.com/oauth2/authorize?client_id=173113690092994561&scope=bot)
to add the official DiscordGo test bot **dgo** to your server. This provides 
indispensable help to this project.

* See [dgVoice](https://github.com/bwmarrin/dgvoice) package for an example of
additional voice helper functions and features for DiscordGo

* See [dca](https://github.com/bwmarrin/dca) for an **experimental** stand alone
tool that wraps `ffmpeg` to create opus encoded audio appropriate for use with
Discord (and DiscordGo)

**For help with this package or general Go discussion, please join the [Discord 
Gophers](https://discord.gg/0f1SbxBZjYq9jLBk) chat server.**

## Getting Started

### master vs develop Branch
* The master branch represents the latest released version of DiscordGo.  This
branch will always have a stable and tested version of the library. Each release
is tagged and you can easily download a specific release and view release notes
on the github [releases](https://github.com/bwmarrin/discordgo/releases) page.

* The develop branch is where all development happens and almost always has
new features over the master branch.  However breaking changes are frequently
added to develop and even sometimes bugs are introduced.  Bugs get fixed and 
the breaking changes get documented before pushing to master.  

*So, what should you use?*

If you can accept the constant changing nature of *develop* then it is the 
recommended branch to use.  Otherwise, if you want to tail behind development
slightly and have a more stable package with documented releases then use *master*

### Installing

This assumes you already have a working Go environment, if not please see
[this page](https://golang.org/doc/install) first.

`go get` *will always pull the latest released version from the master branch.*

```sh
go get github.com/bwmarrin/discordgo
```

If you want to use the develop branch, follow these steps next.

```sh
cd $GOPATH/src/github.com/bwmarrin/discordgo
git checkout develop
```

### Usage

Import the package into your project.

```go
import "github.com/bwmarrin/discordgo"
```

Construct a new Discord client which can be used to access the variety of 
Discord API functions and to set callback functions for Discord events.

```go
discord, err := discordgo.New("Bot " + "authentication token")
```

See Documentation and Examples below for more detailed information.


## Documentation

**NOTICE** : This library and the Discord API are unfinished.
Because of that there may be major changes to library in the future.

The DiscordGo code is fairly well documented at this point and is currently
the only documentation available.  Both GoDoc and GoWalker (below) present
that information in a nice format.

- [![GoDoc](https://godoc.org/github.com/bwmarrin/discordgo?status.svg)](https://godoc.org/github.com/bwmarrin/discordgo) 
- [![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/bwmarrin/discordgo) 
- Hand crafted documentation coming eventually.


## Examples

Below is a list of examples and other projects using DiscordGo.  Please submit 
an issue if you would like your project added or removed from this list 

- [DiscordGo Examples](https://github.com/bwmarrin/discordgo/tree/master/examples) A collection of example programs written with DiscordGo
- [Awesome DiscordGo](https://github.com/bwmarrin/discordgo/wiki/Awesome-DiscordGo) A curated list of high quality projects using DiscordGo

## Troubleshooting
For help with common problems please reference the 
[Troubleshooting](https://github.com/bwmarrin/discordgo/wiki/Troubleshooting) 
section of the project wiki.


## Contributing
Contributions are very welcomed, however please follow the below guidelines.

- First open an issue describing the bug or enhancement so it can be
discussed.  
- Fork the develop branch and make your changes.  
- Try to match current naming conventions as closely as possible.  
- This package is intended to be a low level direct mapping of the Discord API 
so please avoid adding enhancements outside of that scope without first 
discussing it.
- Create a Pull Request with your changes against the develop branch.


## List of Discord APIs

See [this chart](https://abal.moe/Discord/Libraries.html) for a feature 
comparison and list of other Discord API libraries.

## Special Thanks

[Chris Rhodes](https://github.com/iopred) - For the DiscordGo logo and tons of PRs
