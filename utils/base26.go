package utils

var base26Table = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func Base26Encode(str string) uint64 {
	n := uint64(0)
	for _, c := range str {
		n = n*26 + uint64(c-'A')
		n += 1
	}
	return n - 1
}

func Base26Decode(n uint64) string {
	result := ""
	n += 1
	for n > 0 {
		n -= 1
		result = string(base26Table[n%26]) + result
		n /= 26
	}
	return result
}
