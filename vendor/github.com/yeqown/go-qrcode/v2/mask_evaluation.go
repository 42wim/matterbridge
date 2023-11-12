package qrcode

import (
	"math"
)

// evaluation calculate a score after masking matrix.
//
// reference:
// - https://www.thonky.com/qr-code-tutorial/data-masking#Determining-the-Best-Mask
func evaluation(mat *Matrix) (score int) {
	debugLogf("calculate maskScore starting")

	score1 := rule1(mat)
	score2 := rule2(mat)
	score3 := rule3(mat)
	score4 := rule4(mat)
	score = score1 + score2 + score3 + score4
	debugLogf("maskScore: rule1=%d, rule2=%d, rule3=%d, rule4=%d", score1, score2, score3, score4)

	return score
}

// check each row one-by-one. If there are five consecutive modules of the same color,
// add 3 to the penalty. If there are more modules of the same color after the first five,
// add 1 for each additional module of the same color. Afterward, check each column one-by-one,
// checking for the same condition. Add the horizontal and vertical total to obtain penalty score
func rule1(mat *Matrix) (score int) {
	// prerequisites:
	// mat.Width() == mat.Height()
	if mat.Width() != mat.Height() {
		debugLogf("matrix width != height, skip rule1")
		return math.MaxInt32
	}

	dimension := mat.Width()
	scoreLine := func(arr []qrvalue) int {
		lScore, cnt, cur := 0, 0, QRValue_INIT_V0
		for _, v := range arr {
			if !samestate(v, cur) {
				cur = v
				cnt = 1
				continue
			}

			cnt++
			if cnt == 5 {
				lScore += 3
			} else if cnt > 5 {
				lScore++
			}
		}

		return lScore
	}

	for cur := 0; cur < dimension; cur++ {
		row := mat.Row(cur)
		col := mat.Col(cur)
		score += scoreLine(row)
		score += scoreLine(col)
	}

	return score
}

// rule2
// look for areas of the same color that are at least 2x2 modules or larger.
// The QR code specification says that for a solid-color block of size m × n,
// the penalty score is 3 × (m - 1) × (n - 1).
func rule2(mat *Matrix) int {
	var (
		score          int
		s0, s1, s2, s3 qrvalue
	)
	for x := 0; x < mat.Width()-1; x++ {
		for y := 0; y < mat.Height()-1; y++ {
			s0, _ = mat.at(x, y)
			s1, _ = mat.at(x+1, y)
			s2, _ = mat.at(x, y+1)
			s3, _ = mat.at(x+1, y+1)

			if s0 == s1 && s2 == s3 && s1 == s2 {
				score += 3
			}
		}
	}

	return score
}

// rule3 calculate punishment score in rule3, find pattern in QR Code matrix.
// Looks for patterns of dark-light-dark-dark-dark-light-dark that have four
// light modules on either side. In other words, it looks for any of the
// following two patterns: 1011101 0000 or 0000 1011101.
//
// Each time this pattern is found, add 40 to the penalty score.
func rule3(mat *Matrix) (score int) {
	var (
		pattern1     = binaryToQRValueSlice("1011101 0000")
		pattern2     = binaryToQRValueSlice("0000 1011101")
		pattern1Next = kmpGetNext(pattern1)
		pattern2Next = kmpGetNext(pattern2)
	)

	// prerequisites:
	//
	// mat.Width() == mat.Height()
	if mat.Width() != mat.Height() {
		debugLogf("rule3 got matrix but not matched prerequisites")
		return math.MaxInt32
	}
	dimension := mat.Width()

	for i := 0; i < dimension; i++ {
		col := mat.Col(i)
		row := mat.Row(i)

		// DONE(@yeqown): statePattern1 and statePattern2 are fixed, so maybe kmpGetNext
		// could cache result to speed up.
		score += 40 * kmp(col, pattern1, pattern1Next)
		score += 40 * kmp(col, pattern2, pattern2Next)
		score += 40 * kmp(row, pattern1, pattern1Next)
		score += 40 * kmp(row, pattern2, pattern2Next)
	}

	return score
}

// rule4 is based on the ratio of light modules to dark modules:
//
// 1. Count the total number of modules in the matrix.
// 2. Count how many dark modules there are in the matrix.
// 3. Calculate the percent of modules in the matrix that are dark: (darkmodules / totalmodules) * 100
// 4. Determine the previous and next multiple of five of this percent.
// 5. Subtract 50 from each of these multiples of five and take the absolute qrbool of the result.
// 6. Divide each of these by five. For example, 10/5 = 2 and 5/5 = 1.
// 7. Finally, take the smallest of the two numbers and multiply it by 10.
//
func rule4(mat *Matrix) int {
	// prerequisites:
	//
	// mat.Width() == mat.Height()
	if mat.Width() != mat.Height() {
		debugLogf("rule4 got matrix but not matched prerequisites")
		return math.MaxInt32
	}

	dimension := mat.Width()
	dark, total := 0, dimension*dimension
	for i := 0; i < dimension; i++ {
		col := mat.Col(i)

		// count dark modules
		for j := 0; j < dimension; j++ {
			if samestate(col[j], QRValue_DATA_V1) {
				dark++
			}
		}
	}

	ratio := (dark * 100) / total // in range [0, 100]
	step := 0
	if ratio%5 == 0 {
		step = 1
	}

	previous := abs((ratio/5-step)*5 - 50)
	next := abs((ratio/5+1-step)*5 - 50)

	return min(previous, next) / 5 * 10
}
