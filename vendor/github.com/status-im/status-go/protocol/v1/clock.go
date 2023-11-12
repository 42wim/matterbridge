package protocol

import "time"

const clockBumpInMs = uint64(time.Minute / time.Millisecond)

// CalcMessageClock calculates a new clock value for Message.
// It is used to properly sort messages and accommodate the fact
// that time might be different on each device.
func CalcMessageClock(lastObservedValue uint64, timeInMs uint64) uint64 {
	clock := lastObservedValue
	if clock < timeInMs {
		// Added time should be larger than time skew tollerance for a message.
		// Here, we use 1 minute which is larger than accepted message time skew by Whisper.
		clock = timeInMs + clockBumpInMs
	} else {
		clock++
	}
	return clock
}
