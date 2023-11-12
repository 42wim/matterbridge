{.pragma: some, header: "cb.h".}

type
  OnHeaderCallback = proc (s: cstring) {.cdecl.}

proc callbackFn(json: string) {.exportc, cdecl.} =
  echo "callbackFn", json

# callbackFn "some"

proc HelloFromNim(): cstring {.exportc.} = 
    return "Hello, World From Nim\n"

var headerCallback: OnHeaderCallback

proc setHeaderCallback*(cb: OnHeaderCallback) {.exportc.} =
  headerCallback = cb

proc invokeHeaderCallback*() {.exportc.} =
  headerCallback("inside Nim 2222")

proc testEcho*() {.exportc.} =
  echo "in testEcho"




