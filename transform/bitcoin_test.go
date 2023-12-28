package transform

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decentralize-everything/indexer/extract/mempool"
	"github.com/decentralize-everything/indexer/store"
	"go.uber.org/zap"
)

// func TestBitcoinTransformer(t *testing.T) {
// 	logger, _ := zap.NewDevelopment()
// 	btcClient := getblock.NewBitcoinClient("https://go.getblock.io/3c4caebe755b4838838044df4cb08a99")
// 	btcTransformer := NewBitcoinTransformer(logger)
// 	for i := 820000; i < 821000; i++ {
// 		blockHash, err := btcClient.GetBlockHash(i)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		block, err := btcClient.GetBlock(blockHash)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		_, err = btcTransformer.Transform(block)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 	}
// }

func TestBitcoinTransformerWithMempoolClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := store.NewMemDb(false)
	btcClient := mempool.NewBitcoinClient(&chaincfg.MainNetParams)
	btcTransformer := NewBitcoinTransformer(db, logger)
	for i := 820000; i < 820010; i++ {
		blockHash, err := btcClient.GetBlockHash(i)
		if err != nil {
			t.Fatal(err)
		}
		block, err := btcClient.GetBlock(blockHash)
		if err != nil {
			t.Fatal(err)
		}

		_, err = btcTransformer.Transform(block)
		if err != nil {
			t.Fatal(err)
		}
	}
}
