# status-go/protocol

This is an implementation of the [secure transport](https://specs.status.im/spec/5) and [payloads](https://specs.status.im/spec/6) which are a part of [the Status Client specification](https://specs.status.im/spec/1).

This implementation uses SQLite and [SQLCipher](github.com/mutecomm/go-sqlcipher) for persistent storage.

The payloads are encoded using [protocol-buffers](https://developers.google.com/protocol-buffers).

## Content

* `messenger.go` is the main file which exports `Messenger` struct. This is a public API to interact with this implementation of the Status Chat Protocol.
* `protobuf/` contains protobuf files implementing payloads described in [the Payloads spec](https://specs.status.im/spec/6).
* `encryption/` implements [the Secure Transport spec](https://specs.status.im/spec/5).
* `transport/` connects the Status Chat Protocol with a wire-protocol which in our case is either Whisper or Waku.
* `datasync/` is an adapter for [MVDS](https://specs.vac.dev/specs/mvds.html).
* `applicationmetadata/` is an outer layer wrapping a payload with an app-specific metadata like a signature.
* `identity/` implements details related to creating a three-word name and identicon.
* `migrations/` contains implementation specific migrations for the sqlite database which is used by `Messenger` as a persistent data store.

## History

Originally this package was a dedicated repo called `status-protocol-go` and [was migrated](https://github.com/status-im/status-go/pull/1684) into `status-go`. The new `status-go/protocol` package maintained its own dependencies until [sub modules were removed](https://github.com/status-im/status-go/pull/1835/files) and the root go.mod file managed all dependencies for the entire `status-go` repo.   
