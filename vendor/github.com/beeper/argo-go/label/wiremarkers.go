package label

var (
	WireTypeMarkerString     Label
	WireTypeMarkerBoolean    Label
	WireTypeMarkerVarint     Label
	WireTypeMarkerFloat64    Label
	WireTypeMarkerBytes      Label
	WireTypeMarkerFixed      Label
	WireTypeMarkerBlock      Label
	WireTypeMarkerNullable   Label
	WireTypeMarkerArray      Label
	WireTypeMarkerRecord     Label
	WireTypeMarkerDesc       Label
	WireTypeMarkerError      Label
	WireTypeMarkerPath       Label
	WireTypeMarkerUnion      Label
	WireTypeMarkerExtensions Label
)

func init() {
	WireTypeMarkerString = NewFromInt64(-1)
	WireTypeMarkerBoolean = NewFromInt64(-2)
	WireTypeMarkerVarint = NewFromInt64(-3)
	WireTypeMarkerFloat64 = NewFromInt64(-4)
	WireTypeMarkerBytes = NewFromInt64(-5)
	WireTypeMarkerFixed = NewFromInt64(-6)
	WireTypeMarkerBlock = NewFromInt64(-7)
	WireTypeMarkerNullable = NewFromInt64(-8)
	WireTypeMarkerArray = NewFromInt64(-9)
	WireTypeMarkerRecord = NewFromInt64(-10)
	WireTypeMarkerDesc = NewFromInt64(-11)
	WireTypeMarkerError = NewFromInt64(-12)
	WireTypeMarkerPath = NewFromInt64(-13)
	WireTypeMarkerUnion = NewFromInt64(-14)
	WireTypeMarkerExtensions = NewFromInt64(-15)
}
