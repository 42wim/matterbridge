package alias

// For details: https://en.wikipedia.org/wiki/Linear-feedback_shift_register
type LSFR struct {
	data uint64
	poly uint64
}

func newLSFR(poly uint64, seed uint64) *LSFR {
	return &LSFR{data: seed, poly: poly}
}

func (f *LSFR) next() uint64 {
	var bit uint64
	var i uint64

	for i = 0; i < 64; i++ {
		if f.poly&(1<<i) != 0 {
			bit ^= (f.data >> i)
		}
	}
	bit &= 0x01

	f.data = (f.data << 1) | bit

	return f.data
}
