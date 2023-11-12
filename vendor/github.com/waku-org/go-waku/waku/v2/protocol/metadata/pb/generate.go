package pb

//go:generate protoc -I./../../waku-proto/waku/metadata/v1/. -I./../../waku-proto/ --go_opt=paths=source_relative --go_opt=Mwaku_metadata.proto=github.com/waku-org/go-waku/waku/v2/protocol/metadata/pb --go_out=. ./../../waku-proto/waku/metadata/v1/waku_metadata.proto
