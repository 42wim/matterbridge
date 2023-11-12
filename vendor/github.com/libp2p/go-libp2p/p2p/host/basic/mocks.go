//go:build gomock || generate

package basichost

//go:generate sh -c "go run github.com/golang/mock/mockgen -build_flags=\"-tags=gomock\" -package basichost -destination mock_nat_test.go github.com/libp2p/go-libp2p/p2p/host/basic NAT"
type NAT nat
