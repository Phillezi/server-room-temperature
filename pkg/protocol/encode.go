package protocol

func Encode(dst []byte, milliC int32) int {
	if len(dst) < 3 {
		return 0
	}

	i := 0
	dst[i] = '@'
	i++

	value := int64(milliC)

	if value < 0 {
		if i >= len(dst)-1 {
			return 0
		}

		dst[i] = '-'
		i++
		value = -value
	}

	var digits [10]byte
	n := 0

	if value == 0 {
		digits[0] = '0'
		n = 1
	} else {
		for value > 0 {
			digits[n] = byte(value%10) + '0'
			n++
			value /= 10
		}
	}

	if i+n+1 > len(dst) {
		return 0
	}

	for j := n - 1; j >= 0; j-- {
		dst[i] = digits[j]
		i++
	}

	dst[i] = '\n'
	return i + 1
}
