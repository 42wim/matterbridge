## v0.15.0 (2023-03-16)

### beta.3 (2023-03-15)

* **Breaking change *(appservice)*** Removed `Load()` and `AppService.Init()`
  functions. The struct should just be created with `Create()` and the relevant
  fields should be filled manually.
* **Breaking change *(appservice)*** Removed public `HomeserverURL` field and
  replaced it with a `SetHomeserverURL` method.
* *(appservice)* Added support for unix sockets for homeserver URL and
  appservice HTTP server.
* *(client)* Changed request logging to log durations as floats instead of
  strings (using zerolog's `Dur()`, so the exact output can be configured).
* *(bridge)* Changed zerolog to use nanosecond precision timestamps.
* *(crypto)* Added message index to log after encrypting/decrypting megolm
  events, and when failing to decrypt due to duplicate index.
* *(sqlstatestore)* Fixed warning log for rooms that don't have encryption
  enabled.

### beta.2 (2023-03-02)

* *(bridge)* Fixed building with `nocrypto` tag.
* *(bridge)* Fixed legacy logging config migration not disabling file writer
  when `file_name_format` was empty.
* *(bridge)* Added option to require room power level to run commands.
* *(event)* Added structs for [MSC3952]: Intentional Mentions.
* *(util/variationselector)* Added `FullyQualify` method to add necessary emoji
  variation selectors without adding all possible ones.

[MSC3952]: https://github.com/matrix-org/matrix-spec-proposals/pull/3952

### beta.1 (2023-02-24)

* Bumped minimum Go version to 1.19.
* **Breaking changes**
  * *(all)* Switched to zerolog for logging.
    * The `Client` and `Bridge` structs still include a legacy logger for
      backwards compatibility.
  * *(client, appservice)* Moved `SQLStateStore` from appservice module to the
    top-level (client) module.
  * *(client, appservice)* Removed unused `Typing` map in `SQLStateStore`.
  * *(client)* Removed unused `SaveRoom` and `LoadRoom` methods in `Storer`.
  * *(client, appservice)* Removed deprecated `SendVideo` and `SendImage` methods.
  * *(client)* Replaced `AppServiceUserID` field with `SetAppServiceUserID` boolean.
    The `UserID` field is used as the value for the query param.
  * *(crypto)* Renamed `GobStore` to `MemoryStore` and removed the file saving
    features. The data can still be persisted, but the persistence part must be
    implemented separately.
  * *(crypto)* Removed deprecated `DeviceIdentity` alias
    (renamed to `id.Device` long ago).
  * *(client)* Removed `Stringifable` interface as it's the same as `fmt.Stringer`.
* *(client)* Renamed `Storer` interface to `SyncStore`. A type alias exists for
  backwards-compatibility.
* *(crypto/cryptohelper)* Added package for a simplified crypto interface for clients.
* *(example)* Added e2ee support to example using crypto helper.
* *(client)* Changed default syncer to stop syncing on `M_UNKNOWN_TOKEN` errors.

## v0.14.0 (2023-02-16)

* **Breaking change *(format)*** Refactored the HTML parser `Context` to have
  more data.
* *(id)* Fixed escaping path components when forming matrix.to URLs
  or `matrix:` URIs.
* *(bridge)* Bumped default timeouts for decrypting incoming messages.
* *(bridge)* Added `RawArgs` to commands to allow accessing non-split input.
* *(bridge)* Added `ReplyAdvanced` to commands to allow setting markdown
  settings.
* *(event)* Added `notifications` key to `PowerLevelEventContent`.
* *(event)* Changed `SetEdit` to cut off edit fallback if the message is long.
* *(util)* Added `SyncMap` as a simple generic wrapper for a map with a mutex.
* *(util)* Added `ReturnableOnce` as a wrapper for `sync.Once` with a return
  value.

## v0.13.0 (2023-01-16)

* **Breaking change:** Removed `IsTyping` and `SetTyping` in `appservice.StateStore`
  and removed the `TypingStateStore` struct implementing those methods.
* **Breaking change:** Removed legacy fields in Beeper MSS events.
* Added knocked rooms to sync response structs.
* Added wrapper for `/timestamp_to_event` endpoint added in Matrix v1.6.
* Fixed MSC3870 uploads not failing properly after using up the max retry count.
* Fixed parsing non-positive ordered list start positions in HTML parser.

## v0.12.4 (2022-12-16)

* Added `SendReceipt` to support private read receipts and thread receipts in
  the same function. `MarkReadWithContent` is now deprecated.
* Changed media download methods to return errors if the server returns a
  non-2xx status code.
* Removed legacy `sql_store_upgrade.Upgrade` method. Using `store.DB.Upgrade()`
  after `NewSQLCryptoStore(...)` is recommended instead (the bridge module does
  this automatically).
* Added missing `suggested` field to `m.space.child` content struct.
* Added `device_unused_fallback_key_types` to `/sync` response and appservice
  transaction structs.
* Changed `ReqSetReadMarkers` to omit empty fields.
* Changed bridge configs to force `sqlite3-fk-wal` instead of `sqlite3`.
* Updated bridge helper to close database connection when stopping.
* Fixed read receipt and account data endpoints sending `null` instead of an
  empty object as the body when content isn't provided.

## v0.12.3 (2022-11-16)

* **Breaking change:** Added logging for row iteration in the dbutil package.
  This changes the return type of `Query` methods from `*sql.Rows` to a new
  `dbutil.Rows` interface.
* Added flag to disable wrapping database upgrades in a transaction (e.g. to
  allow setting `PRAGMA`s for advanced table mutations on SQLite).
* Deprecated `MessageEventContent.GetReplyTo` in favor of directly using
  `RelatesTo.GetReplyTo`. RelatesTo methods are nil-safe, so checking if
  RelatesTo is nil is not necessary for using those methods.
* Added wrapper for space hierarchyendpoint (thanks to [@mgcm] in [#100]).
* Added bridge config option to handle transactions asynchronously.
* Added separate channels for to-device events in appservice transaction
  handler to avoid blocking to-device events behind normal events.
* Added `RelatesTo.GetNonFallbackReplyTo` utility method to get the reply event
  ID, unless the reply is a thread fallback.
* Added `event.TextToHTML` as an utility method to HTML-escape a string and
  replace newlines with `<br/>`.
* Added check to bridge encryption helper to make sure the e2ee keys are still
  on the server. Synapse is known to sometimes lose keys randomly.
* Changed bridge crypto syncer to crash on `M_UNKNOWN_TOKEN` errors instead of
  retrying forever pointlessly.
* Fixed verifying signatures of fallback one-time keys.

[@mgcm]: https://github.com/mgcm
[#100]: https://github.com/mautrix/go/pull/100

## v0.12.2 (2022-10-16)

* Added utility method to redact bridge commands.
* Added thread ID field to read receipts to match Matrix v1.4 changes.
* Added automatic fetching of media repo config at bridge startup to make it
  easier for bridges to check homeserver media size limits.
* Added wrapper for the `/register/available` endpoint.
* Added custom user agent to all requests mautrix-go makes. The value can be
  customized by changing the `DefaultUserAgent` variable.
* Implemented [MSC3664], [MSC3862] and [MSC3873] in the push rule evaluator.
* Added workaround for potential race conditions in OTK uploads when using
  appservice encryption ([MSC3202]).
* Fixed generating registrations to use `.+` instead of `[0-9]+` in the
  username regex.
* Fixed panic in megolm session listing methods if the store contains withheld
  key entries.
* Fixed missing header in bridge command help messages.

[MSC3664]: https://github.com/matrix-org/matrix-spec-proposals/pull/3664
[MSC3862]: https://github.com/matrix-org/matrix-spec-proposals/pull/3862
[MSC3873]: https://github.com/matrix-org/matrix-spec-proposals/pull/3873

## v0.12.1 (2022-09-16)

* Bumped minimum Go version to 1.18.
* Added `omitempty` for a bunch of fields in response structs to make them more
  usable for server implementations.
* Added `util.RandomToken` to generate GitHub-style access tokens with checksums.
* Added utilities to call the push gateway API.
* Added `unread_notifications` and [MSC2654] `unread_count` fields to /sync
  response structs.
* Implemented [MSC3870] for uploading and downloading media directly to/from an
  external media storage like S3.
* Fixed dbutil database ownership checks on SQLite.
* Fixed typo in unauthorized encryption key withheld code
  (`m.unauthorized` -> `m.unauthorised`).
* Fixed [MSC2409] support to have a separate field for to-device events.

[MSC2654]: https://github.com/matrix-org/matrix-spec-proposals/pull/2654
[MSC3870]: https://github.com/matrix-org/matrix-spec-proposals/pull/3870

## v0.12.0 (2022-08-16)

* **Breaking change:** Switched `Client.UserTyping` to take a `time.Duration`
  instead of raw `int64` milliseconds.
* **Breaking change:** Removed custom reply relation type and switched to using
  the wire format (nesting in `m.in_reply_to`).
* Added device ID to appservice OTK count map to match updated [MSC3202].
  This is also a breaking change, but the previous incorrect behavior wasn't
  implemented by anything other than mautrix-syncproxy/imessage.
* (There are probably other breaking changes too).
* Added database utility and schema upgrade framework
  * Originally from mautrix-whatsapp, but usable for non-bridges too
  * Includes connection wrapper to log query durations and mutate queries for
    SQLite compatibility (replacing `$x` with `?x`).
* Added bridge utilities similar to mautrix-python. Currently includes:
  * Crypto helper
  * Startup flow
  * Command handling and some standard commands
  * Double puppeting things
  * Generic parts of config, basic config validation
  * Appservice SQL state store
* Added alternative markdown spoiler parsing extension that doesn't support
  reasons, but works better otherwise.
* Added Discord underline markdown parsing extension (`_foo_` -> <u>foo</u>).
* Added support for parsing spoilers and color tags in the HTML parser.
* Added support for mutating plain text nodes in the HTML parser.
* Added room version field to the create room request struct.
* Added empty JSON object as default request body for all non-GET requests.
* Added wrapper for `/capabilities` endpoint.
* Added `omitempty` markers for lots of structs to make the structs easier to
  use on the server side too.
* Added support for registering to-device event handlers via the default
  Syncer's `OnEvent` and `OnEventType` methods.
* Fixed `CreateEventContent` using the wrong field name for the room version
  field.
* Fixed `StopSync` not immediately cancelling the sync loop if it was sleeping
  after a failed sync.
* Fixed `GetAvatarURL` always returning the current user's avatar instead of
  the specified user's avatar (thanks to [@nightmared] in [#83]).
* Improved request logging and added new log when a request finishes.
* Crypto store improvements:
  * Deleted devices are now kept in the database.
  * Made ValidateMessageIndex atomic.
* Moved `appservice.RandomString` to the `util` package and made it use
  `crypto/rand` instead of `math/rand`.
* Significantly improved cross-signing validation code.
  * There are now more options for required trust levels,
    e.g. you can set `SendKeysMinTrust` to `id.TrustStateCrossSignedTOFU`
    to trust the first cross-signing master key seen and require all devices
    to be signed by that key.
  * Trust state of incoming messages is automatically resolved and stored in
    `evt.Mautrix.TrustState`. This can be used to reject incoming messages from
    untrusted devices.

[@nightmared]: https://github.com/nightmared
[#83]: https://github.com/mautrix/go/pull/83

## v0.11.1 (2023-01-15)

* Fixed parsing non-positive ordered list start positions in HTML parser
  (backport of the same fix in v0.13.0).

## v0.11.0 (2022-05-16)

* Bumped minimum Go version to 1.17.
* Switched from `/r0` to `/v3` paths everywhere.
  * The new `v3` paths are implemented since Synapse 1.48, Dendrite 0.6.5, and
    Conduit 0.4.0. Servers older than these are no longer supported.
* Switched from blackfriday to goldmark for markdown parsing in the `format`
  module and added spoiler syntax.
* Added `EncryptInPlace` and `DecryptInPlace` methods for attachment encryption.
  In most cases the plain/ciphertext is not necessary after en/decryption, so
  the old `Encrypt` and `Decrypt` are deprecated.
* Added wrapper for `/rooms/.../aliases`.
* Added utility for adding/removing emoji variation selectors to match
  recommendations on reactions in Matrix.
* Added support for async media uploads ([MSC2246]).
* Added automatic sleep when receiving 429 error
  (thanks to [@ownaginatious] in [#44]).
* Added support for parsing spec version numbers from the `/versions` endpoint.
* Removed unstable prefixed constant used for appservice login.
* Fixed URL encoding not working correctly in some cases.

[MSC2246]: https://github.com/matrix-org/matrix-spec-proposals/pull/2246
[@ownaginatious]: https://github.com/ownaginatious
[#44]: https://github.com/mautrix/go/pull/44

## v0.10.12 (2022-03-16)

* Added option to use a different `Client` to send invites in
  `IntentAPI.EnsureJoined`.
* Changed `MessageEventContent` struct to omit empty `msgtype`s in the output
  JSON, as sticker events shouldn't have that field.
* Fixed deserializing the `thumbnail_file` field in `FileInfo`.
* Fixed bug that broke `SQLCryptoStore.FindDeviceByKey`.

## v0.10.11 (2022-02-16)

* Added automatic updating of state store from `IntentAPI` calls.
* Added option to ignore cache in `IntentAPI.EnsureJoined`.
* Added `GetURLPreview` as a wrapper for the `/preview_url` media repo endpoint.
* Moved base58 module inline to avoid pulling in btcd as a dependency.

## v0.10.10 (2022-01-16)

* Added event types and content structs for server ACLs and moderation policy
  lists (thanks to [@qua3k] in [#59] and [#60]).
* Added optional parameter to `Client.LeaveRoom` to pass a `reason` field.

[#59]: https://github.com/mautrix/go/pull/59
[#60]: https://github.com/mautrix/go/pull/60

## v0.10.9 (2022-01-04)

* **Breaking change:** Changed `Messages()` to take a filter as a parameter
  instead of using the syncer's filter (thanks to [@qua3k] in [#55] and [#56]).
  * The previous filter behavior was completely broken, as it sent a whole
    filter instead of just a RoomEventFilter.
  * Passing `nil` as the filter is fine and will disable filtering
    (which is equivalent to what it did before with the invalid filter).
* Added `Context()` wrapper for the `/context` API (thanks to [@qua3k] in [#54]).
* Added utility for converting media files with ffmpeg.

[#54]: https://github.com/mautrix/go/pull/54
[#55]: https://github.com/mautrix/go/pull/55
[#56]: https://github.com/mautrix/go/pull/56
[@qua3k]: https://github.com/qua3k

## v0.10.8 (2021-12-30)

* Added `OlmSession.Describe()` to wrap `olm_session_describe`.
* Added trace logs to log olm session descriptions when encrypting/decrypting
  to-device messages.
* Added space event types and content structs.
* Added support for power level content override field in `CreateRoom`.
* Fixed ordering of olm sessions which would cause an old session to be used in
  some cases even after a client created a new session.

## v0.10.7 (2021-12-16)

* Changed `Client.RedactEvent` to allow arbitrary fields in redaction request.

## v0.10.5 (2021-12-06)

* Fixed websocket disconnection not clearing all pending requests.
* Added `OlmMachine.SendRoomKeyRequest` as a more direct way of sending room
  key requests.
* Added automatic Olm session recreation if an incoming message fails to decrypt.
* Changed `Login` to only omit request content from logs if there's a password
  or token (appservice logins don't have sensitive content).

## v0.10.4 (2021-11-25)

* Added `reason` field to invite and unban requests
  (thanks to [@ptman] in [#48]).
* Fixed `AppService.HasWebsocket()` returning `true` even after websocket has
  disconnected.

[@ptman]: https://github.com/ptman
[#48]: https://github.com/mautrix/go/pull/48

## v0.10.3 (2021-11-18)

* Added logs about incoming appservice transactions.
* Added support for message send checkpoints (as HTTP requests, similar to the
  bridge state reporting system).

## v0.10.2 (2021-11-15)

* Added utility method for finding the first supported login flow matching any
  of the given types.
* Updated registering appservice ghosts to use `inhibit_login` flag to prevent
  lots of unnecessary access tokens from being created.
  * If you want to log in as an appservice ghost, you should use [MSC2778]'s
    appservice login (e.g. like [mautrix-whatsapp does for e2be](https://github.com/mautrix/whatsapp/blob/v0.2.1/crypto.go#L143-L149)).

## v0.10.1 (2021-11-05)

* Removed direct dependency on `pq`
  * In order to use some more efficient queries on postgres, you must set
    `crypto.PostgresArrayWrapper = pq.Array` if you want to use both postgres
    and e2ee.
* Added temporary hack to ignore state events with the MSC2716 historical flag
  (to be removed after [matrix-org/synapse#11265] is merged)
* Added received transaction acknowledgements for websocket appservice
  transactions.
* Added automatic fallback to move `prev_content` from top level to the
  standard location inside `unsigned`.

[matrix-org/synapse#11265]: https://github.com/matrix-org/synapse/pull/11265

## v0.9.31 (2021-10-27)

* Added `SetEdit` utility function for `MessageEventContent`.

## v0.9.30 (2021-10-26)

* Added wrapper for [MSC2716]'s `/batch_send` endpoint.
* Added `MarshalJSON` method for `Event` struct to prevent empty unsigned
  structs from being included in the JSON.

[MSC2716]: https://github.com/matrix-org/matrix-spec-proposals/pull/2716

## v0.9.29 (2021-09-30)

* Added `client.State` method to get full room state.
* Added bridge info structs and event types ([MSC2346]).
* Made response handling more customizable.
* Fixed type of `AuthType` constants.

[MSC2346]: https://github.com/matrix-org/matrix-spec-proposals/pull/2346

## v0.9.28 (2021-09-30)

* Added `X-Mautrix-Process-ID` to appservice websocket headers to help debug
  issues where multiple instances are connecting to the server at the same time.

## v0.9.27 (2021-09-23)

* Fixed Go 1.14 compatibility (broken in v0.9.25).
* Added GitHub actions CI to build, test and check formatting on Go 1.14-1.17.

## v0.9.26 (2021-09-21)

* Added default no-op logger to `Client` in order to prevent panic when the
  application doesn't set a logger.

## v0.9.25 (2021-09-19)

* Disabled logging request JSON for sensitive requests like `/login`,
  `/register` and other UIA endpoints. Logging can still be enabled by
  setting `MAUTRIX_LOG_SENSITIVE_CONTENT` to `yes`.
* Added option to store new homeserver URL from `/login` response well-known data.
* Added option to stream big sync responses via disk to maybe reduce memory usage.
* Fixed trailing slashes in homeserver URL breaking all requests.

## v0.9.24 (2021-09-03)

* Added write deadline for appservice websocket connection.

## v0.9.23 (2021-08-31)

* Fixed storing e2ee key withheld events in the SQL store.

## v0.9.22 (2021-08-30)

* Updated appservice handler to cache multiple recent transaction IDs
  instead of only the most recent one.

## v0.9.21 (2021-08-25)

* Added liveness and readiness endpoints to appservices.
  * The endpoints are the same as mautrix-python:
    `/_matrix/mau/live` and `/_matrix/mau/ready`
  * Liveness always returns 200 and an empty JSON object by default,
    but it can be turned off by setting `appservice.Live` to `false`.
  * Readiness defaults to returning 500, and it can be switched to 200
    by setting `appservice.Ready` to `true`.

## v0.9.20 (2021-08-19)

* Added crypto store migration for converting all `VARCHAR(255)` columns
  to `TEXT` in Postgres databases.

## v0.9.19 (2021-08-17)

* Fixed HTML parser outputting two newlines after paragraph tags.

## v0.9.18 (2021-08-16)

* Added new `BuildURL` method that does the same as `Client.BuildBaseURL`
  but without requiring the `Client` instance.

## v0.9.17 (2021-07-25)

* Fixed handling OTK counts and device lists coming in through the appservice
  transaction websocket.
* Updated OlmMachine to ignore OTK counts intended for other devices.

## v0.9.15 (2021-07-16)

* Added support for [MSC3202] and the to-device part of [MSC2409] in the
  appservice package.
* Added support for sending commands through appservice websocket.
* Changed error message JSON field name in appservice error responses to
  conform with standard Matrix errors (`message` -> `error`).

[MSC3202]: https://github.com/matrix-org/matrix-spec-proposals/pull/3202

## v0.9.14 (2021-06-17)

* Added default implementation of `PillConverter` in HTML parser utility.

## v0.9.13 (2021-06-15)

* Added support for parsing and generating encoded matrix.to URLs and `matrix:` URIs ([MSC2312](https://github.com/matrix-org/matrix-doc/pull/2312)).
* Updated HTML parser to use new URI parser for parsing user/room pills.

## v0.9.12 (2021-05-18)

* Added new method for sending custom data with read receipts
  (not currently a part of the spec).

## v0.9.11 (2021-05-12)

* Improved debug log for unsupported event types.
* Added VoIP events to GuessClass.
* Added support for parsing strings in VoIP event version field.

## v0.9.10 (2021-04-29)

* Fixed `format.RenderMarkdown()` still allowing HTML when both `allowHTML`
  and `allowMarkdown` are `false`.

## v0.9.9 (2021-04-26)

* Updated appservice `StartWebsocket` to return websocket close info.

## v0.9.8 (2021-04-20)

* Added methods for getting room tags and account data.

## v0.9.7 (2021-04-19)

* **Breaking change (crypto):** `SendEncryptedToDevice` now requires an event
  type parameter. Previously it only allowed sending events of type
  `event.ToDeviceForwardedRoomKey`.
* Added content structs for VoIP events.
* Added global mutex for Olm decryption
  (previously it was only used for encryption).

## v0.9.6 (2021-04-15)

* Added option to retry all HTTP requests when encountering a HTTP network
  error or gateway error response (502/503/504)
  * Disabled by default, you need to set the `DefaultHTTPRetries` field in
    the `AppService` or `Client` struct to enable.
  * Can also be enabled with `FullRequest`s `MaxAttempts` field.

## v0.9.5 (2021-04-06)

* Reverted update of `golang.org/x/sys` which broke Go 1.14 / darwin/arm.

## v0.9.4 (2021-04-06)

* Switched appservices to using shared `http.Client` instance with a in-memory
  cookie jar.

## v0.9.3 (2021-03-26)

* Made user agent headers easier to configure.
* Improved logging when receiving weird/unhandled to-device events.

## v0.9.2 (2021-03-15)

* Fixed type of presence state constants (thanks to [@babolivier] in [#30]).
* Implemented presence state fetching methods (thanks to [@babolivier] in [#29]).
* Added support for sending and receiving commands via appservice transaction websocket.

[@babolivier]: https://github.com/babolivier
[#29]: https://github.com/mautrix/go/pull/29
[#30]: https://github.com/mautrix/go/pull/30

## v0.9.1 (2021-03-11)

* Fixed appservice register request hiding actual errors due to UIA error handling.

## v0.9.0 (2021-03-04)

* **Breaking change (manual API requests):** `MakeFullRequest` now takes a
  `FullRequest` struct instead of individual parameters. `MakeRequest`'s
  parameters are unchanged.
* **Breaking change (manual /sync):** `SyncRequest` now requires a `Context`
  parameter.
* **Breaking change (end-to-bridge encryption):**
  the `uk.half-shot.msc2778.login.application_service` constant used for
  appservice login ([MSC2778]) was renamed from `AuthTypeAppservice`
  to `AuthTypeHalfyAppservice`.
  * The `AuthTypeAppservice` constant now contains `m.login.application_service`,
    which is currently only used for registrations, but will also be used for
    login once MSC2778 lands in the spec.
* Fixed appservice registration requests to include `m.login.application_service`
  as the `type` (re [matrix-org/synapse#9548]).
* Added wrapper for `/logout/all`.

[MSC2778]: https://github.com/matrix-org/matrix-spec-proposals/pull/2778
[matrix-org/synapse#9548]: https://github.com/matrix-org/synapse/pull/9548

## v0.8.6 (2021-03-02)

* Added client-side timeout to `mautrix.Client`'s `http.Client`
  (defaults to 3 minutes).
* Updated maulogger to fix bug where plaintext file logs wouldn't have newlines.

## v0.8.5 (2021-02-26)

* Fixed potential concurrent map writes in appservice `Client` and `Intent`
  methods.

## v0.8.4 (2021-02-24)

* Added option to output appservice logs as JSON.
* Added new methods for validating user ID localparts.

## v0.8.3 (2021-02-21)

* Allowed empty content URIs in parser
* Added functions for device management endpoints
  (thanks to [@edwargix] in [#26]).

[@edwargix]: https://github.com/edwargix
[#26]: https://github.com/mautrix/go/pull/26

## v0.8.2 (2021-02-09)

* Fixed error when removing the user's avatar.

## v0.8.1 (2021-02-09)

* Added AccountDataStore to remove the need for persistent local storage other
  than the access token (thanks to [@daenney] in [#23]).
* Added support for receiving appservice transactions over websocket.
  See <https://github.com/mautrix/wsproxy> for the server-side implementation.
* Fixed error when removing the room avatar.

[@daenney]: https://github.com/daenney
[#23]: https://github.com/mautrix/go/pull/23

## v0.8.0 (2020-12-24)

* **Breaking change:** the `RateLimited` field in the `Registration` struct is
  now a pointer, so that it can be omitted entirely.
* Merged initial SSSS/cross-signing code by [@nikofil]. Interactive verification
  doesn't work, but the other things mostly do.
* Added support for authorization header auth in appservices ([MSC2832]).
* Added support for receiving ephemeral events directly ([MSC2409]).
* Fixed `SendReaction()` and other similar methods in the `Client` struct.
* Fixed crypto cgo code panicking in Go 1.15.3+.
* Fixed olm session locks sometime getting deadlocked.

[MSC2832]: https://github.com/matrix-org/matrix-spec-proposals/pull/2832
[MSC2409]: https://github.com/matrix-org/matrix-spec-proposals/pull/2409
[@nikofil]: https://github.com/nikofil
