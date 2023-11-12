package qrcode

// samestate judge two matrix qrtype is same with binary semantic.
// QRValue_DATA_V0/QRType_INIT only equal to QRValue_DATA_V0, other state are equal to each other.
func samestate(s1, s2 qrvalue) bool {
	return s1.qrbool() == s2.qrbool()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func min(x, y int) int {
	if x < y {
		return x
	}

	return y
}

func binaryToQRValueSlice(s string) []qrvalue {
	var states = make([]qrvalue, 0, len(s))
	for _, c := range s {
		switch c {
		case '1':
			states = append(states, QRValue_DATA_V1)
		case '0':
			states = append(states, QRValue_DATA_V0)
		default:
			continue
		}
	}
	return states
}
