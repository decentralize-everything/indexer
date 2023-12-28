package getblock

import "testing"

func TestBitcoinClient(t *testing.T) {
	client := NewBitcoinClient("https://go.getblock.io/3c4caebe755b4838838044df4cb08a99")
	// blockCount, err := client.GetBlockCount()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log(blockCount)

	// blockHash, err := client.GetLatestBlockHash()
	blockHash, err := client.GetBlockHash(822823)
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

func TestGetLatestBlockHeight(t *testing.T) {
	client := NewBitcoinClient("https://go.getblock.io/3c4caebe755b4838838044df4cb08a99")
	blockCount, err := client.GetLatestBlockHeight()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(blockCount)
}
