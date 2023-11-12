package pb

//go:generate protoc -I./../../waku-proto/waku/peer_exchange/v2alpha1/. -I./../../waku-proto/ --go_opt=paths=source_relative --go_opt=Mpeer_exchange.proto=github.com/waku-org/go-waku/waku/v2/protocol/peer_exchange/pb --go_out=. ./../../waku-proto/waku/peer_exchange/v2alpha1/peer_exchange.proto
