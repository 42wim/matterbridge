<p align="center"><a href="https://pkg.go.dev/github.com/lrstanley/girc"><img width="270" src="http://i.imgur.com/DEnyrdB.png"></a></p>
<!-- template:begin:header -->
<!-- do not edit anything in this "template" block, its auto-generated -->

<p align="center">girc -- :bomb: girc is a flexible IRC library for Go :ok_hand:</p>
<p align="center">
  <a href="https://github.com/lrstanley/girc/tags">
    <img title="Latest Semver Tag" src="https://img.shields.io/github/v/tag/lrstanley/girc?style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/girc/commits/master">
    <img title="Last commit" src="https://img.shields.io/github/last-commit/lrstanley/girc?style=flat-square">
  </a>

  <a href="https://github.com/lrstanley/girc/actions?query=workflow%3Atest+event%3Apush">
    <img title="GitHub Workflow Status (test @ master)" src="https://img.shields.io/github/actions/workflow/status/lrstanley/girc/test.yml?branch=master&label=test&style=flat-square">
  </a>

  <a href="https://codecov.io/gh/lrstanley/girc">
    <img title="Code Coverage" src="https://img.shields.io/codecov/c/github/lrstanley/girc/master?style=flat-square">
  </a>

  <a href="https://pkg.go.dev/github.com/lrstanley/girc">
    <img title="Go Documentation" src="https://pkg.go.dev/badge/github.com/lrstanley/girc?style=flat-square">
  </a>
  <a href="https://goreportcard.com/report/github.com/lrstanley/girc">
    <img title="Go Report Card" src="https://goreportcard.com/badge/github.com/lrstanley/girc?style=flat-square">
  </a>
</p>
<p align="center">
  <a href="https://github.com/lrstanley/girc/issues?q=is:open+is:issue+label:bug">
    <img title="Bug reports" src="https://img.shields.io/github/issues/lrstanley/girc/bug?label=issues&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/girc/issues?q=is:open+is:issue+label:enhancement">
    <img title="Feature requests" src="https://img.shields.io/github/issues/lrstanley/girc/enhancement?label=feature%20requests&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/girc/pulls">
    <img title="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lrstanley/girc?label=prs&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/girc/discussions/new?category=q-a">
    <img title="Ask a Question" src="https://img.shields.io/badge/support-ask_a_question!-blue?style=flat-square">
  </a>
  <a href="https://liam.sh/chat"><img src="https://img.shields.io/badge/discord-bytecord-blue.svg?style=flat-square" title="Discord Chat"></a>
</p>
<!-- template:end:header -->

<!-- template:begin:toc -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :link: Table of Contents

  - [Features](#features)
  - [Installing](#installing)
  - [Examples](#examples)
  - [References](#references)
  - [Support &amp; Assistance](#raising_hand_man-support--assistance)
  - [Contributing](#handshake-contributing)
  - [License](#balance_scale-license)
<!-- template:end:toc -->

## Features

- Focuses on simplicity, yet tries to still be flexible.
- Only requires [standard library packages](https://godoc.org/github.com/lrstanley/girc?imports)
- Event based triggering/responses ([example](https://godoc.org/github.com/lrstanley/girc#ex-package--Commands), and [CTCP too](https://godoc.org/github.com/lrstanley/girc#Commands.SendCTCP)!)
- [Documentation](https://godoc.org/github.com/lrstanley/girc) is _mostly_ complete.
- Support for a good portion of the [IRCv3 spec](http://ircv3.net/software/libraries.html).
  - SASL Auth (currently only `PLAIN` and `EXTERNAL` is support by default,
  however you can simply implement `SASLMech` yourself to support additional
  mechanisms.)
  - Message tags (things like `account-tag` on by default)
  - `account-notify`, `away-notify`, `chghost`, `extended-join`, etc -- all handled seemlessly ([cap.go](https://github.com/lrstanley/girc/blob/master/cap.go) for more info).
- Channel and user tracking. Easily find what users are in a channel, if a
  user is away, or if they are authenticated (if the server supports it!)
- Client state/capability tracking. Easy methods to access capability data ([LookupChannel](https://godoc.org/github.com/lrstanley/girc#Client.LookupChannel), [LookupUser](https://godoc.org/github.com/lrstanley/girc#Client.LookupUser), [GetServerOption (ISUPPORT)](https://godoc.org/github.com/lrstanley/girc#Client.GetServerOption), etc.)
- Built-in support for things you would commonly have to implement yourself.
  - Nick collision detection and prevention (also see [Config.HandleNickCollide](https://godoc.org/github.com/lrstanley/girc#Config).)
  - Event/message rate limiting.
  - Channel, nick, and user validation methods ([IsValidChannel](https://godoc.org/github.com/lrstanley/girc#IsValidChannel), [IsValidNick](https://godoc.org/github.com/lrstanley/girc#IsValidNick), etc.)
  - CTCP handling and auto-responses ([CTCP](https://godoc.org/github.com/lrstanley/girc#CTCP))
  - And more!

## Installing

    $ go get -u github.com/lrstanley/girc

## Examples

See [the examples](https://godoc.org/github.com/lrstanley/girc#example-package--Bare)
within the documentation for real-world usecases. Here are a few real-world
usecases/examples/projects which utilize girc:

| Project | Description |
| --- | --- |
| [nagios-check-ircd](https://github.com/lrstanley/nagios-check-ircd) | Nagios utility for monitoring the health of an ircd |
| [nagios-notify-irc](https://github.com/lrstanley/nagios-notify-irc) | Nagios utility for sending alerts to one or many channels/networks |
| [matterbridge](https://github.com/42wim/matterbridge) | bridge between mattermost, IRC, slack, discord (and many others) with REST API |

Working on a project and want to add it to the list? Submit a pull request!

## References

   * [IRCv3: Specification Docs](http://ircv3.net/irc/)
   * [IRCv3: Specification Repo](https://github.com/ircv3/ircv3-specifications)
   * [IRCv3 Capability Registry](http://ircv3.net/registry.html)
   * [IRCv3: WEBIRC](https://ircv3.net/specs/extensions/webirc.html)
   * [KiwiIRC: WEBIRC](https://kiwiirc.com/docs/webirc)
   * [ISUPPORT Specification Docs](http://www.irc.org/tech_docs/005.html) ([alternative 1](http://defs.ircdocs.horse/defs/isupport.html), [alternative 2](https://github.com/grawity/irc-docs/blob/master/client/RPL_ISUPPORT/draft-hardy-irc-isupport-00.txt), [relevant draft](http://www.irc.org/tech_docs/draft-brocklesby-irc-isupport-03.txt))
   * [IRC Numerics List](http://defs.ircdocs.horse/defs/numerics.html)
   * [Extended WHO (also known as WHOX)](https://github.com/quakenet/snircd/blob/master/doc/readme.who)
   * [RFC1459: Internet Relay Chat Protocol](https://tools.ietf.org/html/rfc1459)
   * [RFC2812: Internet Relay Chat: Client Protocol](https://tools.ietf.org/html/rfc2812)
   * [RFC2813: Internet Relay Chat: Server Protocol](https://tools.ietf.org/html/rfc2813)
   * [RFC7194: Default Port for Internet Relay Chat (IRC) via TLS/SSL](https://tools.ietf.org/html/rfc7194)
   * [RFC4422: Simple Authentication and Security Layer](https://tools.ietf.org/html/rfc4422) ([SASL EXTERNAL](https://tools.ietf.org/html/rfc4422#appendix-A))
   * [RFC4616: The PLAIN SASL Mechanism](https://tools.ietf.org/html/rfc4616)


<!-- template:begin:support -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :raising_hand_man: Support & Assistance

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for
     guidelines on ensuring everyone has the best experience interacting with
     the community.
* :raising_hand_man: Take a look at the [support](.github/SUPPORT.md) document on
     guidelines for tips on how to ask the right questions.
* :lady_beetle: For all features/bugs/issues/questions/etc, [head over here](https://github.com/lrstanley/girc/issues/new/choose).
<!-- template:end:support -->

<!-- template:begin:contributing -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :handshake: Contributing

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for guidelines
     on ensuring everyone has the best experience interacting with the
    community.
* :clipboard: Please review the [contributing](.github/CONTRIBUTING.md) doc for submitting
     issues/a guide on submitting pull requests and helping out.
* :old_key: For anything security related, please review this repositories [security policy](https://github.com/lrstanley/girc/security/policy).
<!-- template:end:contributing -->

<!-- template:begin:license -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :balance_scale: License

```
MIT License

Copyright (c) 2016 Liam Stanley <me@liamstanley.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

_Also located [here](LICENSE)_
<!-- template:end:license -->
girc artwork licensed under [CC 3.0](http://creativecommons.org/licenses/by/3.0/)
based on Renee French under Creative Commons 3.0 Attributions.
