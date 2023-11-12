package rendezvous_pb

//go:generate protoc -I. --proto_path=./ --go_opt=paths=source_relative --go_opt=Mrendezvous.proto=github.com/waku-org/go-libp2p-rendezvous/rendezvous_pb --go_out=. ./rendezvous.proto
