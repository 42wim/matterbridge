# torrent

[![PkgGoDev](https://pkg.go.dev/badge/github.com/anacrolix/torrent)](https://pkg.go.dev/github.com/anacrolix/torrent)

This repository implements BitTorrent-related packages and command-line utilities in Go. The emphasis is on use as a library from other projects. It's been used 24/7 in production by downstream services since late 2014. The implementation was specifically created to explore Go's concurrency capabilities, and to include the ability to stream data directly from the BitTorrent network. To this end it [supports seeking, readaheads and other features](https://godoc.org/github.com/anacrolix/torrent#Reader) exposing torrents and their files with the various Go idiomatic `io` package interfaces. This is also demonstrated through [torrentfs](#torrentfs).

There is [support for protocol encryption, DHT, PEX, uTP, and various extensions](https://godoc.org/github.com/anacrolix/torrent). There are [several data storage backends provided](https://godoc.org/github.com/anacrolix/torrent/storage): blob, file, bolt, mmap, and sqlite, to name a few. You can [write your own](https://godoc.org/github.com/anacrolix/torrent/storage#ClientImpl) to store data for example on S3, or in a database. 

Some noteworthy package dependencies that can be used for other purposes include:

 * [go-libutp](https://github.com/anacrolix/go-libutp)
 * [dht](https://github.com/anacrolix/dht)
 * [bencode](https://godoc.org/github.com/anacrolix/torrent/bencode)
 * [tracker](https://godoc.org/github.com/anacrolix/torrent/tracker)

## Installation

Install the library package with `go get github.com/anacrolix/torrent`, or the provided cmds with `go install github.com/anacrolix/torrent/cmd/...@latest`.

## Library examples

There are some small [examples](https://godoc.org/github.com/anacrolix/torrent#pkg-examples) in the package documentation.

## Mentions

 * [@anacrolix](https://github.com/anacrolix) is interviewed about this repo in [Console 32](https://console.substack.com/p/console-32).

### Downstream projects

There are several web-frontends and Android clients among the known public projects:

 * [Torrent.Express](https://torrent.express/)
 * [Confluence](https://github.com/anacrolix/confluence)
 * [Trickl](https://github.com/arranlomas/Trickl)
 * [Elementum](http://elementum.surge.sh/) (up to version 0.0.71)
 * [goTorrent](https://github.com/deranjer/goTorrent)
 * [Go Peerflix](https://github.com/Sioro-Neoku/go-peerflix)
 * [Simple Torrent](https://github.com/boypt/simple-torrent) (fork of [Cloud Torrent](https://github.com/jpillora/cloud-torrent), unmaintained)
 * [Android Torrent Client](https://gitlab.com/axet/android-torrent-client)
 * [libtorrent](https://gitlab.com/axet/libtorrent)
 * [Remote-Torrent](https://github.com/BruceWangNo1/remote-torrent)
 * [ANT-Downloader](https://github.com/anatasluo/ant)
 * [Go-PeersToHTTP](https://github.com/WinPooh32/peerstohttp)
 * [CortexFoundation/torrentfs](https://github.com/CortexFoundation/torrentfs): P2P file system of cortex full node
 * [TorrServ](https://github.com/YouROK/TorrServer): Torrent streaming server over http.
 * [Webtor](https://webtor.io/): free cloud BitTorrent-client

## Help

Communication about the project is primarily through [Discussions](https://github.com/anacrolix/torrent/discussions) and the [issue tracker](https://github.com/anacrolix/torrent/issues).

## Command packages

Here I'll describe what some of the packages in `./cmd` do. See [installation](#installation) to make them available.

### torrent

#### torrent download

Downloads torrents from the command-line.

    $ torrent download 'magnet:?xt=urn:btih:KRWPCX3SJUM4IMM4YF5RPHL6ANPYTQPU'
    ... lots of jibba jabber ...
    downloading "ubuntu-14.04.2-desktop-amd64.iso": 1.0 GB/1.0 GB, 1989/1992 pieces completed (1 partial)
    2015/04/01 02:08:20 main.go:137: downloaded ALL the torrents
    $ md5sum ubuntu-14.04.2-desktop-amd64.iso
    1b305d585b1918f297164add46784116  ubuntu-14.04.2-desktop-amd64.iso
    $ echo such amaze
    wow

#### torrent metainfo magnet

Creates a magnet link from a torrent file. Note the extracted trackers, display name, and info hash.

    $ torrent metainfo testdata/debian-10.8.0-amd64-netinst.iso.torrent magnet
    magnet:?xt=urn:btih:4090c3c2a394a49974dfbbf2ce7ad0db3cdeddd7&dn=debian-10.8.0-amd64-netinst.iso&tr=http%3A%2F%2Fbttracker.debian.org%3A6969%2Fannounce

See `torrent metainfo --help` for other metainfo related commands.

### torrentfs

torrentfs mounts a FUSE filesystem at `-mountDir`. The contents are the torrents described by the torrent files and magnet links at `-metainfoDir`. Data for read requests is fetched only as required from the torrent network, and stored at `-downloadDir`.

    $ mkdir mnt torrents
    $ torrentfs -mountDir=mnt -metainfoDir=torrents &
    $ cd torrents
    $ wget http://releases.ubuntu.com/14.04.2/ubuntu-14.04.2-desktop-amd64.iso.torrent
    $ cd ..
    $ ls mnt
    ubuntu-14.04.2-desktop-amd64.iso
    $ pv mnt/ubuntu-14.04.2-desktop-amd64.iso | md5sum
    996MB 0:04:40 [3.55MB/s] [========================================>] 100%
    1b305d585b1918f297164add46784116  -

