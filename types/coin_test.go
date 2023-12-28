package types

import (
	"reflect"
	"testing"
)

func TestCoinInfoCodec(t *testing.T) {
	ci := &CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max": uint64(100),
		},
		TxCount:     1,
		HolderCount: 1,
		CreatedAt:   2,
	}

	bs := ci.ToBytes()
	ci2 := &CoinInfo{}
	if err := ci2.FromBytes(bs); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ci, ci2) {
		t.Fatal("not equal")
	}
}

func TestUnspentCoinCodec(t *testing.T) {
	uc := &UnspentCoin{
		CoinId: "CARV",
		Owner:  "1234",
		Amount: 1,
		Utxo:   "5678",
	}

	bs := uc.ToBytes()
	uc2 := &UnspentCoin{}
	if err := uc2.FromBytes(bs); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(uc, uc2) {
		t.Fatal("not equal")
	}
}
