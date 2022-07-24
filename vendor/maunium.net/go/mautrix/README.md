# mautrix-go
[![GoDoc](https://pkg.go.dev/badge/maunium.net/go/mautrix)](https://pkg.go.dev/maunium.net/go/mautrix)

A Golang Matrix framework. Used by [gomuks](https://matrix.org/docs/projects/client/gomuks),
[go-neb](https://github.com/matrix-org/go-neb), [mautrix-whatsapp](https://github.com/mautrix/whatsapp)
and others.

Matrix room: [`#maunium:maunium.net`](https://matrix.to/#/#maunium:maunium.net)

This project is based on [matrix-org/gomatrix](https://github.com/matrix-org/gomatrix).
The original project is licensed under [Apache 2.0](https://github.com/matrix-org/gomatrix/blob/master/LICENSE).

In addition to the basic client API features the original project has, this framework also has:

* Appservice support (Intent API like mautrix-python, room state storage, etc)
* End-to-end encryption support (incl. interactive SAS verification)
* Structs for parsing event content
* Helpers for parsing and generating Matrix HTML
* Helpers for handling push rules

This project contains modules that are licensed under Apache 2.0:

* [maunium.net/go/mautrix/crypto/canonicaljson](crypto/canonicaljson)
* [maunium.net/go/mautrix/crypto/olm](crypto/olm)
