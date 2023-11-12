package hop

//go:generate abigen --abi l2SaddleSwap.abi --pkg hopSwap --out swap/l2SaddleSwap.go
//go:generate abigen --abi l1Bridge.abi --pkg hopBridge --out bridge/l1Bridge.go
//go:generate abigen --abi l2AmmWrapper.abi --pkg hopWrapper --out wrapper/l2AmmWrapper.go
