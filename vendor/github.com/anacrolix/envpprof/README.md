# envpprof

[![pkg.go.dev badge](https://img.shields.io/badge/pkg.go.dev-reference-blue)](https://pkg.go.dev/github.com/anacrolix/envpprof)

Allows run-time configuration of Go's pprof features and default HTTP mux using the environment variable `GOPPROF`. Import the package with `import _ "github.com/anacrolix/envpprof"`. `envpprof` has an `init` function that will run at process initialization that checks the value of the `GOPPROF` environment variable. The variable can contain a comma-separated list of values, for example `GOPPROF=http,block`. The supported keys are:

Key | Effect
--- | ------
http | Exposes the default HTTP muxer [`"net/http".DefaultServeMux`](https://pkg.go.dev/net/http?tab=doc#pkg-variables) to the first free TCP port after `6060` on `localhost`. The process PID, and location are logged automatically when this is enabled. `DefaultServeMux` is frequently the default location to expose status, and debugging endpoints, including those provided by [`net/http/pprof`](https://pkg.go.dev/net/http/pprof?tab=doc). Note that the `net/http/pprof` import is included with `envpprof`, and exposed on `DefaultServeMux`.
cpu |Calls [`"runtime/pprof".StartCPUProfile`](https://pkg.go.dev/runtime/pprof?tab=doc#StartCPUProfile), writing to a temporary file in `$HOME/pprof` with the prefix `cpu`. The file is not removed after use. The name of the file is logged when this is enabled. [`envpprof.Stop`](https://pkg.go.dev/github.com/anacrolix/envpprof?tab=doc#Stop) should be deferred from `main` when this will be used, to ensure proper clean up.
heap |This is similar to the `cpu` key, but writes heap profile information to a file prefixed with `heap`. The profile will not be written unless `Stop` is invoked. See `cpu` for more. 
block | This calls [`"runtime".SetBlockProfileRate(1)`](https://pkg.go.dev/runtime?tab=doc#SetBlockProfileRate) enabling the profiling of goroutine blocking events. Note that if `http` is enabled, this exposes the blocking profile at the HTTP path `/debug/pprof/block` per package [`net/http/pprof`](https://pkg.go.dev/net/http/pprof?tab=doc#pkg-overview).
mutex | This calls [`"runtime".SetMutexProfileFraction(1)`](https://pkg.go.dev/runtime?tab=doc#SetMutexProfileFraction) enabling profiling of mutex contention events. Note that if `http` is enabled, this exposes the profile at the HTTP path `/debug/pprof/mutex` per package [`net/http/pprof`](https://pkg.go.dev/net/http/pprof?tab=doc#pkg-overview).

