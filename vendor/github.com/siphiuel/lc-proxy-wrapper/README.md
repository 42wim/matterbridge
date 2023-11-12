# Build Nim
nim c --app:staticlib --header:cb.h --noMain:on --nimcache:$HOME/c/lc-proxy-wrapper/nimcache cb.nim

# Build go
go build
./lc-proxy-wrapper
