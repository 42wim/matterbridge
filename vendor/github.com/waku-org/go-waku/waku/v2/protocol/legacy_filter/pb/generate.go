package pb

//go:generate mv ./../../waku-proto/waku/filter/v2beta1/filter.proto ./../../waku-proto/waku/filter/v2beta1/legacy_filter.proto
//go:generate protoc -I./../../waku-proto/waku/filter/v2beta1/. -I./../../waku-proto/ --go_opt=paths=source_relative --go_opt=Mlegacy_filter.proto=github.com/waku-org/go-waku/waku/v2/protocol/legacy_filter/pb --go_opt=Mwaku/message/v1/message.proto=github.com/waku-org/go-waku/waku/v2/protocol/pb --go_out=. ./../../waku-proto/waku/filter/v2beta1/legacy_filter.proto
//go:generate mv ./../../waku-proto/waku/filter/v2beta1/legacy_filter.proto ./../../waku-proto/waku/filter/v2beta1/filter.proto
