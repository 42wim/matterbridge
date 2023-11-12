release: release-osx release-linux

release-osx:
	- mkdir -p draft/osx
	GOOS=darwin GOARCH=amd64 go build -o draft/osx/go-qrcode ./cmd/go-qrcode.go
	cd draft/osx && tar -zcvf ../go-qrcode.osx.tar.gz .

release-linux:
	- mkdir -p draft/linux
	GOOS=linux GOARCH=amd64 go build -o draft/linux/go-qrcode ./cmd/go-qrcode.go
	cd draft/linux && tar -zcvf ../go-qrcode.linux.tar.gz .

test-all:
	go test -v --count=1 ./...