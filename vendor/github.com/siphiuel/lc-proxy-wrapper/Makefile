include Makefile.vars

build-verif-proxy-wrapper:
	CGO_CFLAGS="$(CGO_CFLAGS)" go build -x -v -ldflags $(LDFLAGS)

build-verif-proxy-wrapper-exe:
	CGO_CFLAGS="$(CGO_CFLAGS)" go build -x -v -ldflags $(LDFLAGS) -o verif-proxy-wrapper ./main 

.PHONY: clean

clean:
	rm -rf nimcache libcb.a verif-proxy-wrapper

