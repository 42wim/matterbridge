package pb

//go:generate protoc -I./../../waku-proto/waku/store/v2beta4//. -I./../../waku-proto/ --go_opt=paths=source_relative --go_opt=Mstore.proto=github.com/waku-org/go-waku/waku/v2/protocol/store/pb --go_opt=Mwaku/message/v1/message.proto=github.com/waku-org/go-waku/waku/v2/protocol/pb --go_out=. ./../../waku-proto/waku/store/v2beta4/store.proto
