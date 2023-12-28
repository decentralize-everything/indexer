package mempool

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestBitcoinClient(t *testing.T) {
	client := NewBitcoinClient(&chaincfg.MainNetParams)
	blockHash, err := client.GetBlockHash(823122)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(blockHash)

	block, err := client.GetBlock(blockHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(block)
}
