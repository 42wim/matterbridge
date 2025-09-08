# Building from source

This page documents how to build matterbridge from source. If you're looking for ready-to-use executables, head over to [setup.md].

If you really want to build from source, follow these instructions:
Go 1.18+ is required. Make sure you have [Go](https://golang.org/doc/install) properly installed.

Building the binary with **all** the bridges enabled needs about 3GB RAM to compile.
You can reduce this memory requirement to 0,5GB RAM by adding the `nomsteams` tag if you don't need/use the Microsoft Teams bridge.

Matterbridge can be build without gcc/c-compiler: If you're running on windows first run `set CGO_ENABLED=0` on other platforms you prepend `CGO_ENABLED=0` to the `go build` command. (eg `CGO_ENABLED=0 go install github.com/42wim/matterbridge`)

To install the latest stable run:

```bash
go install github.com/42wim/matterbridge
```

To install the latest dev run:

```bash
go install github.com/42wim/matterbridge@master
```

To install the latest stable run without msteams or zulip bridge:

```bash
go install -tags nomsteams,nozulip github.com/42wim/matterbridge
```

You should now have matterbridge binary in the ~/go/bin directory:

```bash
$ ls ~/go/bin/
matterbridge
```

## Building with whatsapp (beta) multidevice support

Because the library we use for Whatsapp multidevice support includes a GPL3 library we can not provide you binaries.
(as this would require the Matterbridge to change it license to GPL)

Matterbridge can be build without gcc/c-compiler: If you're running on windows first run `set CGO_ENABLED=0` on other platforms you prepend `CGO_ENABLED=0` to the `go build` command. (eg `CGO_ENABLED=0 go install github.com/42wim/matterbridge`)

So this means you have to build it yourself using the instructions below:

```bash
go install -tags whatsappmulti github.com/42wim/matterbridge@master
```

If you're low on memory and don't need msteams:

```bash
go install -tags nomsteams,whatsappmulti github.com/42wim/matterbridge@master
```

You should now have matterbridge binary in the ~/go/bin directory:

```bash
$ ls ~/go/bin/
matterbridge
```
