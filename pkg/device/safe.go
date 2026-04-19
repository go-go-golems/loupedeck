package device

import "fmt"

func checkedIntToByte(name string, v int) (byte, error) {
	if v < 0 || v > 255 {
		return 0, fmt.Errorf("%s=%d out of byte range", name, v)
	}
	return byte(v), nil
}

func checkedButtonToByte(name string, b Button) (byte, error) {
	if b > 255 {
		return 0, fmt.Errorf("%s=%d out of byte range", name, b)
	}
	return byte(b), nil
}

func clampIntToUint16(v int) uint16 {
	if v <= 0 {
		return 0
	}
	if v >= 65535 {
		return 65535
	}
	return uint16(v)
}
