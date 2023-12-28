package utils

import "testing"

func TestBase26Codec(t *testing.T) {
	str := "PSBTS"
	n := Base26Encode(str)
	if n != 7647450 {
		t.Errorf("base26 encode error: %d", n)
	}

	str = Base26Decode(n)
	if str != "PSBTS" {
		t.Errorf("base26 decode error: %s", str)
	}
}

func TestBase26EncodeCARV(t *testing.T) {
	str := "CARV"
	n := Base26Encode(str)
	t.Log(n)

	str = "INVALID"
	n = Base26Encode(str)
	t.Log(n)
}
