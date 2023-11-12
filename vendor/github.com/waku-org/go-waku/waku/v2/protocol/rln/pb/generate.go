package pb

//go:generate protoc -I./../../waku-proto/waku/rln/v1/. -I./../../waku-proto/ --go_opt=paths=source_relative --go_opt=Mrln.proto=github.com/waku-org/go-waku/waku/v2/protocol/rln/pb --go_out=. ./../../waku-proto/waku/rln/v1/rln.proto
