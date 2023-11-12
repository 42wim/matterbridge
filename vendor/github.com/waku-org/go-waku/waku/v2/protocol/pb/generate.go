package pb

//go:generate protoc -I./../waku-proto/waku/message/v1/. -I./../waku-proto/ --go_opt=paths=source_relative --go_opt=Mmessage.proto=github.com/waku-org/go-waku/waku/v2/pb --go_out=. ./../waku-proto/waku/message/v1/message.proto
