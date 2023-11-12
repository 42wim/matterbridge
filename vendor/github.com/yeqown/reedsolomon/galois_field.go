package reedsolomon

// ref to https://en.wikiversity.org/wiki/Reed%E2%80%93Solomon_codes_for_coders
// ref to https://www.jianshu.com/p/8208aad537bb
// gf.go Galois Fields

var (
	gfLog = []byte{}
	gfExp = []byte{}
)

const prim = 0x011d

// init calls all initial funcs
func init() {
	initTables()
}

// init gfExp and gfLog array
func initTables() {
	gfExp = make([]byte, 512)
	gfLog = make([]byte, 256)

	var (
		x uint16 = 1
	)

	for i := 0; i < 255; i++ {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)

		x <<= 1
		// x overflow 256
		if (x & 0x100) != 0 {
			x ^= prim
		}
	}

	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

// multpy
func gfMul(x, y byte) byte {
	if x == 0 || y == 0 {
		return 0
	}
	// byte max: 256 but exp cap is 512
	return gfExp[uint(gfLog[x])+uint(gfLog[y])]
}

// divide
// func gfDiv(x, y byte) byte {
// 	if y == 0 {
// 		panic("zero division error")
// 	}
// 	if x == 0 {
// 		return 0
// 	}
// 	return gfExp[(uint(gfLog[x])+255-uint(gfLog[y]))%255]
// }

// // inverse
// func gfInverse(x byte) byte {
// 	return gfExp[255-uint(gfLog[x])]
// }

// // pow
// func gfPow(x, power byte) byte {
// 	return gfExp[(gfLog[x]*power)%255]
// }
