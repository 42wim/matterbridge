package pb

//go:generate protoc -I./../../waku-proto/waku/lightpush/v2beta1/. -I./../../waku-proto/ --go_opt=paths=source_relative --go_opt=Mlightpush.proto=github.com/waku-org/go-waku/waku/v2/protocol/lightpush/pb --go_opt=Mwaku/message/v1/message.proto=github.com/waku-org/go-waku/waku/v2/protocol/pb --go_out=. ./../../waku-proto/waku/lightpush/v2beta1/lightpush.proto
