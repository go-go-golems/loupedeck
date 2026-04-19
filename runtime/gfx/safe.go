package gfx

func clampIntToUint8(v int) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v)
}

func clampUint32ToUint8(v uint32) uint8 {
	if v >= 255 {
		return 255
	}
	return uint8(v)
}

func scaleUint8(a, brightness uint8) uint8 {
	product := int(a) * int(brightness)
	return clampIntToUint8(product / 255)
}
