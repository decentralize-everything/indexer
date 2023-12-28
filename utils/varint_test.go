package utils

import (
	"encoding/binary"
	"testing"
)

func TestVarintDecode(t *testing.T) {
	i := 0
	// data := []byte{0x02, 0x82, 0xd1, 0xe0, 0x5a, 0x86, 0x68, 0x80, 0x3c}
	// 6a 01 52 0b 00 bf80c793db8f52 8668 00
	// 6a 01 52 0b 00 80c793db8f52 80e930 02
	data := []byte{0x00, 0xbf, 0x80, 0xc7, 0x93, 0xdb, 0x8f, 0x52, 0x86, 0x68, 0x00}
	for {
		v, l := VarintDecode(data[i:])
		i += l
		t.Logf("%d", v)
		if len(data) == i {
			return
		}
	}
}

func TestVarintEncode(t *testing.T) {
	t.Logf("%x", VarintEncode(2))
	t.Logf("%x", VarintEncode(7647450))
	t.Logf("%x", VarintEncode(1000))
	t.Logf("%x", VarintEncode(188))
}

func TestVarintEncodeArray(t *testing.T) {
	out := VarintEncodeArray([]uint64{53893, 21000000, 10000})
	t.Logf("%x", out)
}

func TestGolangVarint(t *testing.T) {
	out := make([]byte, 80)
	i := 0
	i += binary.PutUvarint(out[i:], 2)
	i += binary.PutUvarint(out[i:], 7647450)
	i += binary.PutUvarint(out[i:], 1000)
	i += binary.PutUvarint(out[i:], 188)
	t.Log(out[:i])

	bs := out[:i]
	pos := 0
	n, len := binary.Uvarint(bs[pos:])
	t.Log(n)
	pos += len
	n, len = binary.Uvarint(bs[pos:])
	t.Log(n)
	pos += len
	n, len = binary.Uvarint(bs[pos:])
	t.Log(n)
	pos += len
	n, len = binary.Uvarint(bs[pos:])
	t.Log(n)
	pos += len
	t.Log(pos)
}
