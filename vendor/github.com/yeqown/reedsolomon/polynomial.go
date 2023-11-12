package reedsolomon

// generator polynomial
// (x-a^1) * (x - a^2) * .... * (x -a^numECWords-1)
func rsGenPoly(numECWords int) []byte {
	var generator = []byte{1}
	for i := 0; i < numECWords; i++ {
		generator = polyMul(generator, []byte{1, gfExp[i]})
	}
	return generator
}

// 将一个多项式和一个标量相乘
func polyScale(poly []byte, x byte) []byte {
	result := make([]byte, len(poly))
	for i := 0; i < len(poly); i++ {
		result[i] = gfMul(poly[i], x)
	}
	return result
}

func polyAdd(poly1, poly2 []byte) []byte {
	size1 := len(poly1)
	size2 := len(poly2)
	size := size1

	if size2 > size1 {
		size = size2
	}
	result := make([]byte, size)

	for i := 0; i < size1; i++ {
		result[i] = byte(poly1[i])
	}

	for i := 0; i < size2; i++ {
		result[i] ^= byte(poly2[i])
	}
	return result
}

// mul polynomial
func polyMul(poly1, poly2 []byte) []byte {
	result := make([]byte, len(poly1)+len(poly2)-1)
	for i := 0; i < len(poly1); i++ {
		for j := 0; j < len(poly2); j++ {
			result[i+j] ^= gfMul(poly1[i], poly2[j])
		}
	}
	return result
}

// func polyEval(poly []byte, x byte) byte {
// 	y := poly[0]
// 	for i := 1; i < len(poly); i++ {
// 		y = gfMul(y, x) ^ poly[i]
// 	}

// 	return y
// }

// ref to: https://www.thonky.com/qr-code-tutorial/show-division-steps?msg_coeff=12%2C34%2C56%2C23&num_ecc_blocks=3
func polyDiv(dividend, divisor []byte) []byte {
	if len(dividend) == 0 {
		panic("could not div with 0 length dividend")
	}

	var (
		leadTerm       = dividend[0]
		reminder, a, b []byte
	)

	reminder = dividend

	for i := 0; i < len(dividend); i++ {
		// step a: generator * leadTerm
		a = polyScale(divisor, leadTerm)

		// step b, xor operation
		b = polyAdd(reminder, a)

		// discard lead term of b
		reminder = b[1:]
		leadTerm = reminder[0]
	}

	return reminder
}
