package utils

// func VarintEncode(n uint64) []byte {
// 	if n == 0 {
// 		return []byte{0}
// 	}

// 	result := []byte{}
// 	for n > 0 {
// 		b := byte(n % 128)
// 		if len(result) > 0 {
// 			b += 127
// 		}
// 		result = append([]byte{b}, result...)
// 		n /= 128
// 	}
// 	return result
// }

func VarintEncode(n uint64) []byte {
	result := [18]byte{}
	i := 17
	result[i] = byte(n & 0x7f)
	for n > 0x7f {
		n = n/128 - 1
		i -= 1
		result[i] = byte(n | 0x80)
	}
	return result[i:]
}

func VarintDecode(data []byte) (uint64, int) {
	v := uint64(0)
	i := 0
	for {
		v = v * 128
		b := data[i]
		if b < 128 {
			return v + uint64(b), i + 1
		} else {
			v = v + uint64(b-127)
			i++
		}
	}
}

func VarintEncodeArray(arr []uint64) []byte {
	result := []byte{}
	for _, n := range arr {
		result = append(result, VarintEncode(n)...)
	}
	return result
}

func VarintDecodeArray(data []byte) []uint64 {
	result := []uint64{}
	i := 0
	for {
		v, l := VarintDecode(data[i:])
		result = append(result, v)
		i += l
		if len(data) == i {
			return result
		}
	}
}
