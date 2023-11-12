package krpc

type Infohash [20]byte

type CompactInfohashes [][20]byte

func (CompactInfohashes) ElemSize() int { return 20 }

func (me CompactInfohashes) MarshalBinary() ([]byte, error) {
	return marshalBinarySlice(me)
}

func (me CompactInfohashes) MarshalBencode() ([]byte, error) {
	return bencodeBytesResult(me.MarshalBinary())
}

func (me *CompactInfohashes) UnmarshalBinary(b []byte) error {
	return unmarshalBinarySlice(me, b)
}

func (me *CompactInfohashes) UnmarshalBencode(b []byte) error {
	return unmarshalBencodedBinary(me, b)
}
