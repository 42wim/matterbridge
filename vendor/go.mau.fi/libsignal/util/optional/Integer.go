package optional

// NewOptionalUint32 returns an optional Uint32 structure.
func NewOptionalUint32(value uint32) *Uint32 {
	return &Uint32{Value: value, IsEmpty: false}
}

func NewEmptyUint32() *Uint32 {
	return &Uint32{IsEmpty: true}
}

// Uint32 is a simple structure for Uint32 values that can
// optionally be nil.
type Uint32 struct {
	Value   uint32
	IsEmpty bool
}
