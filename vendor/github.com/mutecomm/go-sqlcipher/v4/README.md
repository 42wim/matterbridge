## go-sqlcipher

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/mutecomm/go-sqlcipher) [![CI](https://github.com/mutecomm/go-sqlcipher/workflows/CI/badge.svg)](https://github.com/mutecomm/go-sqlcipher/actions)

### Description

Self-contained Go sqlite3 driver with an AES-256 encrypted sqlite3 database
conforming to the built-in database/sql interface. It is based on:

- Go sqlite3 driver: https://github.com/mattn/go-sqlite3
- SQLite extension with AES-256 codec: https://github.com/sqlcipher/sqlcipher
- AES-256 implementation from: https://github.com/libtom/libtomcrypt

SQLite itself is part of SQLCipher.

### Incompatibilities of SQLCipher

The version tags of go-sqlcipher are the same as for SQLCipher.

**SQLCipher 4.x is incompatible with SQLCipher 3.x!**

go-sqlcipher does not implement any migration strategies at the moment.
So if you upgrade a major version of go-sqlcipher, you yourself are responsible
to upgrade existing database files.

See [migrating databases](https://www.zetetic.net/sqlcipher/sqlcipher-api/#Migrating_Databases) for details.

To upgrade your Go code to the 4.x series, change the import path to

    "github.com/mutecomm/go-sqlcipher/v4"

### Installation

This package can be installed with the go get command:

    go get github.com/mutecomm/go-sqlcipher


### Documentation

To create and open encrypted database files use the following DSN parameters:

```go
key := "2DD29CA851E7B56E4697B0E1F08507293D761A05CE4D1B628663F411A8086D99"
dbname := fmt.Sprintf("db?_pragma_key=x'%s'&_pragma_cipher_page_size=4096", key)
db, _ := sql.Open("sqlite3", dbname)
```

`_pragma_key` is the hex encoded 32 byte key (must be 64 characters long).
`_pragma_cipher_page_size` is the page size of the encrypted database (set if
you want a different value than the default size).

```go
key := url.QueryEscape("secret")
dbname := fmt.Sprintf("db?_pragma_key=%s&_pragma_cipher_page_size=4096", key)
db, _ := sql.Open("sqlite3", dbname)
```

This uses a passphrase directly as `_pragma_key` with the key derivation function in
SQLCipher. Do not forget the `url.QueryEscape()` call in your code!

See also [PRAGMA key](https://www.zetetic.net/sqlcipher/sqlcipher-api/#PRAGMA_key).

API documentation can be found here:
http://godoc.org/github.com/mutecomm/go-sqlcipher

Use the function
[sqlite3.IsEncrypted()](https://godoc.org/github.com/mutecomm/go-sqlcipher#IsEncrypted)
to check whether a database file is encrypted or not.

Examples can be found under the `./_example` directory


### License

The code of the originating packages is covered by their respective licenses.
See [LICENSE](LICENSE) file for details.
