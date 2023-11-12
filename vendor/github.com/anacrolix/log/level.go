package log

import (
	"strconv"
)

type Level struct {
	rank   int
	logStr string
}

var levelKey = new(struct{})

var (
	NotSet   = Level{0, "UNSET"}
	Debug    = Level{1, "DEBUG"}
	Info     = Level{2, "INFO"}
	Warning  = Level{3, "WARN"}
	Error    = Level{4, "ERROR"}
	Critical = Level{5, "CRIT"}
	// Will this get special treatment? Not yet.
	Fatal = Level{6, "FATAL"}
)

func (l Level) LogString() string {
	switch l.rank {
	case NotSet.rank:
		return "unset"
	case Debug.rank:
		return "debug"
	case Info.rank:
		return "info"
	case Warning.rank:
		return "warn"
	case Error.rank:
		return "error"
	case Critical.rank:
		return "crit"
	case Fatal.rank:
		return "fatal"
	default:
		return strconv.FormatInt(int64(l.rank), 10)
	}
}

func (l Level) LessThan(r Level) bool {
	if l.rank == NotSet.rank {
		return false
	}
	return l.rank < r.rank
}
