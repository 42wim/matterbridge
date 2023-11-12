# `zombiezen.com/go/sqlite` Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

[Unreleased]: https://github.com/zombiezen/go-sqlite/compare/v0.8.0...main

## [0.8.0][] - 2021-11-07

Version 0.8 adds new transaction functions to `sqlitex`.

[0.8.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.8.0

### Added

- Added `sqlitex.Transaction`, `sqlitex.ImmediateTransaction`, and
  `sqlitex.ExclusiveTransaction`.

## [0.7.2][] - 2021-09-11

[0.7.2]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.7.2

### Fixed

- Updated `modernc.org/sqlite` dependency to a released version instead of a
  prerelease

## [0.7.1][] - 2021-09-09

[0.7.1]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.7.1

### Added

- Added an example to `sqlitemigration.Schema`

## [0.7.0][] - 2021-08-27

[0.7.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.7.0

### Added

- `sqlitemigration.Schema` has a new option for disabling foreign keys for
  individual migrations. This makes it easier to perform migrations that require
  [reconstructing a table][]. ([#20](https://github.com/zombiezen/go-sqlite/issues/20))

[reconstructing a table]: https://sqlite.org/lang_altertable.html#making_other_kinds_of_table_schema_changes

### Changed

- `sqlitemigration.Migrate` and `*sqlitemigration.Pool` no longer use a
  transaction to apply the entire set of migrations: they now only use
  transactions during each individual migration. This was never documented, so
  in theory no one should be depending on this behavior. However, this does mean
  that two processes trying to open and migrate a database concurrently may race
  to apply migrations, whereas before only one process would acquire the write
  lock and migrate.

### Fixed

- Fixed compile breakage on 32-bit architectures. Thanks to Jan Mercl for the
  report.

## [0.6.2][] - 2021-08-17

[0.6.2]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.6.2

### Changed

- `*sqlitex.Pool.Put` now accepts `nil` instead of panicing.
  ([#17](https://github.com/zombiezen/go-sqlite/issues/17))

## [0.6.1][] - 2021-08-16

[0.6.1]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.6.1

### Fixed

- Fixed a potential memory corruption issue introduced in 0.6.0. Thanks to
  Jan Mercl for the report.

## [0.6.0][] - 2021-08-15

[0.6.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.6.0

### Added

- Added back the session API: `Session`, `ChangesetIterator`, `Changegroup`, and
  various functions. There are some slight naming changes from the
  `crawshaw.io/sqlite` API, but they can all be migrated automatically with the
  migration tool. ([#16](https://github.com/zombiezen/go-sqlite/issues/16))

### Changed

- Method calls to a `nil` `*sqlite.Conn` will return an error rather than panic.
  ([#17](https://github.com/zombiezen/go-sqlite/issues/17))

### Removed

- Removed `OpenFlags` that are only used for VFS.

### Fixed

- Properly clean up WAL when using `sqlitex.Pool`
  ([#14](https://github.com/zombiezen/go-sqlite/issues/14))
- Disabled double-quoted string literals.

## [0.5.0][] - 2021-05-22

[0.5.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.5.0

### Added

- Added `shell` package with basic [REPL][]
- Added `SetAuthorizer`, `Limit`, and `SetDefensive` methods to `*Conn` for use
  in ([#12](https://github.com/zombiezen/go-sqlite/issues/12))
- Added `Version` and `VersionNumber` constants

[REPL]: https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop

### Fixed

- Documented compiled-in extensions ([#11](https://github.com/zombiezen/go-sqlite/issues/11))
- Internal objects are no longer susceptible to ID wraparound issues
  ([#13](https://github.com/zombiezen/go-sqlite/issues/13))

## [0.4.0][] - 2021-05-13

[0.4.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.4.0

### Added

- Add Context.Conn method ([#10](https://github.com/zombiezen/go-sqlite/issues/10))
- Add methods to get and set auxiliary function data
  ([#3](https://github.com/zombiezen/go-sqlite/issues/3))

## [0.3.1][] - 2021-05-03

[0.3.1]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.3.1

### Fixed

- Fix conversion of BLOB to TEXT when returning BLOB from a user-defined function

## [0.3.0][] - 2021-04-27

[0.3.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.3.0

### Added

- Implement `io.StringWriter`, `io.ReaderFrom`, and `io.WriterTo` on `Blob`
  ([#2](https://github.com/zombiezen/go-sqlite/issues/2))
- Add godoc examples for `Blob`, `sqlitemigration`, and `SetInterrupt`
- Add more README documentation

## [0.2.2][] - 2021-04-24

[0.2.2]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.2.2

### Changed

- Simplified license to [ISC](https://github.com/zombiezen/go-sqlite/blob/v0.2.2/LICENSE)

### Fixed

- Updated version of `modernc.org/sqlite` to 1.10.4 to use [mutex initialization](https://gitlab.com/cznic/sqlite/-/issues/52)
- Fixed doc comment for `BindZeroBlob`

## [0.2.1][] - 2021-04-17

[0.2.1]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.2.1

### Fixed

- Removed bogus import comment

## [0.2.0][] - 2021-04-03

[0.2.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.2.0

### Added

- New migration tool. See [the README](https://github.com/zombiezen/go-sqlite/blob/v0.2.0/cmd/zombiezen-sqlite-migrate/README.md)
  to get started. ([#1](https://github.com/zombiezen/go-sqlite/issues/1))

### Changed

- `*Conn.CreateFunction` has changed entirely. See the
  [reference](https://pkg.go.dev/zombiezen.com/go/sqlite#Conn.CreateFunction)
  for details.
- `sqlitex.File` and `sqlitex.Buffer` have been moved to the `sqlitefile` package
- The `sqlitefile.Exec*` functions have been moved to the `sqlitex` package
  as `Exec*FS`.

## [0.1.0][] - 2021-03-31

Initial release

[0.1.0]: https://github.com/zombiezen/go-sqlite/releases/tag/v0.1.0
