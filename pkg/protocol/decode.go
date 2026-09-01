package protocol

import "strconv"

func Decode(frame []byte) (int32, bool) {
	if len(frame) < 3 {
		return 0, false
	}

	if frame[0] != '@' || frame[len(frame)-1] != '\n' {
		return 0, false
	}

	// Exclude '@' and '\n'.
	value, err := strconv.ParseInt(
		string(frame[1:len(frame)-1]),
		10,
		32,
	)
	if err != nil {
		return 0, false
	}

	return int32(value), true
}
