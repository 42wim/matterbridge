package qrcode

type QRType = qrtype

// qrtype
type qrtype uint8

const (
	// QRType_INIT represents the initial block state of the matrix
	QRType_INIT qrtype = 1 << 1
	// QRType_DATA represents the data block state of the matrix
	QRType_DATA qrtype = 2 << 1
	// QRType_VERSION indicates the version block of matrix
	QRType_VERSION qrtype = 3 << 1
	// QRType_FORMAT indicates the format block of matrix
	QRType_FORMAT qrtype = 4 << 1
	// QRType_FINDER indicates the finder block of matrix
	QRType_FINDER qrtype = 5 << 1
	// QRType_DARK ...
	QRType_DARK     qrtype = 6 << 1
	QRType_SPLITTER qrtype = 7 << 1
	QRType_TIMING   qrtype = 8 << 1
)

func (s qrtype) String() string {
	switch s {
	case QRType_INIT:
		return "I"
	case QRType_DATA:
		return "d"
	case QRType_VERSION:
		return "V"
	case QRType_FORMAT:
		return "f"
	case QRType_FINDER:
		return "F"
	case QRType_DARK:
		return "D"
	case QRType_SPLITTER:
		return "S"
	case QRType_TIMING:
		return "T"
	}

	return "?"
}

type QRValue = qrvalue

func (v QRValue) Type() qrtype {
	return v.qrtype()
}

func (v QRValue) IsSet() bool {
	return v.qrbool()
}

// qrvalue represents the value of the matrix, it is composed of the qrtype(7bits) and the value(1bits).
// such as: 0b0000,0011 (QRValue_DATA_V1) represents the qrtype is QRType_DATA and the value is 1.
type qrvalue uint8

var (
	// QRValue_INIT_V0 represents the value 0
	QRValue_INIT_V0 = qrvalue(QRType_INIT | 0)

	// QRValue_DATA_V0 represents the block has been set to false
	QRValue_DATA_V0 = qrvalue(QRType_DATA | 0)
	// QRValue_DATA_V1 represents the block has been set to TRUE
	QRValue_DATA_V1 = qrvalue(QRType_DATA | 1)

	// QRValue_VERSION_V0 represents the block has been set to false
	QRValue_VERSION_V0 = qrvalue(QRType_VERSION | 0)
	// QRValue_VERSION_V1 represents the block has been set to TRUE
	QRValue_VERSION_V1 = qrvalue(QRType_VERSION | 1)

	// QRValue_FORMAT_V0 represents the block has been set to false
	QRValue_FORMAT_V0 = qrvalue(QRType_FORMAT | 0)
	// QRValue_FORMAT_V1 represents the block has been set to TRUE
	QRValue_FORMAT_V1 = qrvalue(QRType_FORMAT | 1)

	// QRValue_FINDER_V0 represents the block has been set to false
	QRValue_FINDER_V0 = qrvalue(QRType_FINDER | 0)
	// QRValue_FINDER_V1 represents the block has been set to TRUE
	QRValue_FINDER_V1 = qrvalue(QRType_FINDER | 1)

	// QRValue_DARK_V0 represents the block has been set to false
	QRValue_DARK_V0 = qrvalue(QRType_DARK | 0)
	// QRValue_DARK_V1 represents the block has been set to TRUE
	QRValue_DARK_V1 = qrvalue(QRType_DARK | 1)

	// QRValue_SPLITTER_V0 represents the block has been set to false
	QRValue_SPLITTER_V0 = qrvalue(QRType_SPLITTER | 0)
	// QRValue_SPLITTER_V1 represents the block has been set to TRUE
	QRValue_SPLITTER_V1 = qrvalue(QRType_SPLITTER | 1)

	// QRValue_TIMING_V0 represents the block has been set to false
	QRValue_TIMING_V0 = qrvalue(QRType_TIMING | 0)
	// QRValue_TIMING_V1 represents the block has been set to TRUE
	QRValue_TIMING_V1 = qrvalue(QRType_TIMING | 1)
)

func (v qrvalue) qrtype() qrtype {
	return qrtype(v & 0xfe)
}

func (v qrvalue) qrbool() bool {
	return v&0x01 == 1
}

func (v qrvalue) String() string {
	t := v.qrtype()
	if v.qrbool() {
		return t.String() + "1"
	}

	return t.String() + "0"
}

func (v qrvalue) xor(v2 qrvalue) qrvalue {
	if v != v2 {
		return QRValue_DATA_V1
	}

	return QRValue_DATA_V0
}
