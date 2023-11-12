package pb

//go:generate protoc -I./../../waku-proto/waku/filter/v2/. -I./../../waku-proto/ --go_opt=paths=source_relative --go_opt=Mfilter.proto=github.com/waku-org/go-waku/waku/v2/protocol/filter/pb --go_opt=Mwaku/message/v1/message.proto=github.com/waku-org/go-waku/waku/v2/protocol/pb --go_out=. ./../../waku-proto/waku/filter/v2/filter.proto
