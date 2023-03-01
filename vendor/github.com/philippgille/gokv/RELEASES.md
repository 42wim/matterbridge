Releases
========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

vNext
-----

- Added support for [Go modules](https://github.com/golang/go/wiki/Modules) (issue [#81](https://github.com/philippgille/gokv/issues/81))
    - All `gokv.Store` implementations are now separate Go modules
- Added `gokv.Store` implementations:
    - Package `hazelcast` - A `gokv.Store` implementation for [Hazelcast](https://github.com/hazelcast/hazelcast) (issue [#75](https://github.com/philippgille/gokv/issues/75))
- Fixed: Compile error in `badgerdb` after a breaking change in BadgerDB 1.6.0

v0.5.0 (2019-01-12)
-------------------

- Added: Package `encoding` - An abstraction and wrapper for the core functionality of packages like `encoding/json` and `encoding/gob` (issue [#47](https://github.com/philippgille/gokv/issues/47))
- Added: Package `sql` - It contains shared code for SQL implementations. `mysql` and `postgres` already use it and if you want to create your own SQL implementation you can use it as well. (Useful for issue [#57](https://github.com/philippgille/gokv/issues/57).)
- Added `gokv.Store` implementations:
    - Package `s3` - A `gokv.Store` implementation for [Amazon S3](https://aws.amazon.com/s3/) (issue [#37](https://github.com/philippgille/gokv/issues/37))
        - Also works for other S3-compatible cloud services like [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) and [Scaleway Object Storage](https://www.scaleway.com/object-storage/), as well as for self-hosted solutions like [OpenStack Swift](https://github.com/openstack/swift), [Ceph](https://github.com/ceph/ceph) and [Minio](https://github.com/minio/minio)
    - Package `tablestorage` - A `gokv.Store` implementation for [Azure Table Storage](https://azure.microsoft.com/en-us/services/storage/tables/) (issue [#42](https://github.com/philippgille/gokv/issues/42))
    - Package `datastore` - A `gokv.Store` implementation for [Google Cloud Datastore](https://cloud.google.com/datastore/) (issue [#51](https://github.com/philippgille/gokv/issues/51))
    - Package `tablestore` - A `gokv.Store` implementation for [Alibaba Cloud Table Store](https://www.alibabacloud.com/de/product/table-store) (issue [#70](https://github.com/philippgille/gokv/issues/70))
    - Package `leveldb` - A `gokv.Store` implementation for [LevelDB](https://github.com/syndtr/goleveldb) (issue [#48](https://github.com/philippgille/gokv/issues/48))
    - Package `file` - A `gokv.Store` implementation for storing key-value pairs as files (issue [#52](https://github.com/philippgille/gokv/issues/52))
    - Package `zookeeper` - A `gokv.Store` implementation for [Apache ZooKeeper](https://github.com/apache/zookeeper) (issue [#66](https://github.com/philippgille/gokv/issues/66))
    - Package `postgresql` - A `gokv.Store` implementation for [PostgreSQL](https://github.com/postgres/postgres) (issue [#57](https://github.com/philippgille/gokv/issues/57))
    - Package `cockroachdb` - A `gokv.Store` implementation for [CockroachDB](https://github.com/cockroachdb/cockroach) (issue [#62](https://github.com/philippgille/gokv/issues/62))
    - Package `ignite` - A `gokv.Store` implementation for [Apache Ignite](https://github.com/apache/ignite) (issue [#64](https://github.com/philippgille/gokv/issues/64))
    - Package `freecache` - A `gokv.Store` implementation for [FreeCache](https://github.com/coocood/freecache) (issue [#44](https://github.com/philippgille/gokv/issues/44))
    - Package `bigcache` - A `gokv.Store` implementation for [BigCache](https://github.com/allegro/bigcache) (issue [#45](https://github.com/philippgille/gokv/issues/45))

Breaking changes
----------------

- The `MarshalFormat` enums were removed from all packages that contained `gokv.Store` implementations. Instead the shared package `encoding` was introduced (required for issue [#47](https://github.com/philippgille/gokv/issues/47))

v0.4.0 (2018-12-02)
-------------------

- Added: Method `Close() error` (issue [#36](https://github.com/philippgille/gokv/issues/36))
- Added `gokv.Store` implementations:
    - Package `mongodb` - A `gokv.Store` implementation for [MongoDB](https://github.com/mongodb/mongo) (issue [#27](https://github.com/philippgille/gokv/issues/27))
    - Package `dynamodb` - A `gokv.Store` implementation for [Amazon DynamoDB](https://aws.amazon.com/dynamodb/) (issue [#28](https://github.com/philippgille/gokv/issues/28))
    - Package `memcached` - A `gokv.Store` implementation for [Memcached](https://github.com/memcached/memcached) (issue [#31](https://github.com/philippgille/gokv/issues/31))
    - Package `mysql` - A `gokv.Store` implementation for [MySQL](https://github.com/mysql/mysql-server) (issue [#32](https://github.com/philippgille/gokv/issues/32))
- Added: The factory function `redis.NewClient()` now checks if the connection to the Redis server works and otherwise returns an error.
- Added: The `test` package now has the function `func TestConcurrentInteractions(t *testing.T, goroutineCount int, store gokv.Store)` that you can use to test your `gokv.Store` implementation with concurrent interactions.
- Improved: The `etcd.Client` timeout implementation was improved.
- Fixed: The `Get()` method of the `bbolt` store ignored errors if they occurred during the retrieval of the value
- Fixed: Spelling in error message when using the etcd implementation and the etcd server is unreachable

### Breaking changes

- The added `Close() error` method (see above) means that previous implementations of `gokv.Store` are not compatible with the interface anymore.
- Renamed `bolt` package to `bbolt` to reflect the fact that the maintained fork is used. Also changed all other occurrences of "bolt" (e.g. in GoDoc comments etc.).
- Due to the above mentioned addition to the Redis client factory function, the function signature changed from `func NewClient(options Options) Client` to `func NewClient(options Options) (Client, error)`.

v0.3.0 (2018-11-17)
-------------------

- Added: Method `Delete(string) error` (issue [#8](https://github.com/philippgille/gokv/issues/8))
- Added: All `gokv.Store` implementations in this package now also support [gob](https://blog.golang.org/gobs-of-data) as marshal format as alternative to JSON (issue [#22](https://github.com/philippgille/gokv/issues/22))
    - Part of this addition are a new field in the existing `Options` structs, called `MarshalFormat`, as well as the related `MarshalFormat` enum (custom type + related `const` values) in each implementation package
- Added `gokv.Store` implementations:
    - Package `badgerdb` - A `gokv.Store` implementation for [BadgerDB](https://github.com/dgraph-io/badger) (issue [#16](https://github.com/philippgille/gokv/issues/16))
    - Package `consul` - A `gokv.Store` implementation for [Consul](https://github.com/hashicorp/consul) (issue [#18](https://github.com/philippgille/gokv/issues/18))
    - Package `etcd` - A `gokv.Store` implementation for [etcd](https://github.com/etcd-io/etcd) (issue [#24](https://github.com/philippgille/gokv/issues/24))

### Breaking changes

- The added `Delete(string) error` method (see above) means that previous implementations of `gokv.Store` are not compatible with the interface anymore.
- Changed: The `NewStore()` function in `gomap` and `syncmap` now has an `Option` parameter. Required for issue [#22](https://github.com/philippgille/gokv/issues/22).
- Changed: Passing an empty string as key to `Set()`, `Get()` or `Delete()` now results in an error
- Changed: Passing `nil` as value parameter to `Set()` or as pointer to `Get()` now results in an error. This change leads to a consistent behaviour across the different marshal formats (otherwise for example `encoding/json` marshals `nil` to `null` while `encoding/gob` returns an error).

v0.2.0 (2018-11-05)
-------------------

- Added `gokv.Store` implementation:
    - Package `gomap` - A `gokv.Store` implementation for a plain Go map with a `sync.RWMutex` for concurrent access (issue [#11](https://github.com/philippgille/gokv/issues/11))
- Improved: Every `gokv.Store` implementation resides in its own package now, so when downloading the package of an implementation, for example with `go get github.com/philippgille/gokv/redis`, only the actually required dependencies are downloaded and compiled, making the process much faster. This is especially useful for example when creating Docker images, where in many cases (depending on the `Dockerfile`) the download and compilation are repeated for *each build*. (Issue [#2](https://github.com/philippgille/gokv/issues/2))
- Improved: The performance of `bolt.Store` should be higher, because unnecessary manual locking was removed. (Issue [#1](https://github.com/philippgille/gokv/issues/1))
- Fixed: The `gokv.Store` implementation for bbolt / Bolt DB used data from within a Bolt transaction outside of it, without copying the value, which can lead to errors (see [here](https://github.com/etcd-io/bbolt/blob/76a4670663d125b6b89d47ea3cc659a282d87c28/doc.go#L38)) (issue [#13](https://github.com/philippgille/gokv/issues/13))

### Breaking changes

- All `gokv.Store` implementations were moved into their own packages and the structs that implement the interface were renamed to avoid unidiomatic "stuttering".

v0.1.0 (2018-10-14)
-------------------

Initial release with code from [philippgille/ln-paywall:78fd1dfbf10f549a22f4f30ac7f68c2a2735e989](https://github.com/philippgille/ln-paywall/tree/78fd1dfbf10f549a22f4f30ac7f68c2a2735e989) with only a few changes like a different default path and a bucket name as additional option for the Bolt DB implementation.

Features:

- Interface with `Set(string, interface{}) error` and `Get(string, interface{}) (bool, error)`
- Implementations for:
    - [bbolt](https://github.com/etcd-io/bbolt) (formerly known as Bolt / Bolt DB)
    - Go map (`sync.Map`)
    - [Redis](https://github.com/antirez/redis)
