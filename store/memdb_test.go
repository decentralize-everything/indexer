package store

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"

	"github.com/decentralize-everything/indexer/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMemDbBalanceBatchUpdate(t *testing.T) {
	db := NewMemDb("", "testnet", false, nil)
	db.coins["c1"] = &types.CoinInfo{}
	db.coins["c2"] = &types.CoinInfo{}
	db.coinAddressBalance["c1"] = map[string]int{
		"a1": 1,
		"a2": 2,
	}
	db.addressCoinBalance["a1"] = map[string]int{
		"c1": 1,
	}
	db.addressCoinBalance["a2"] = map[string]int{
		"c1": 2,
	}

	db.BalanceBatchUpdate(map[string]map[string]int{
		"c1": {
			"a1": -1,
			"a2": 1,
		},
		"c2": {
			"a1": 1,
		},
	})

	assert.Equal(t, db.coinAddressBalance["c1"]["a2"], 3)
	assert.Equal(t, db.coinAddressBalance["c2"]["a1"], 1)
	assert.Equal(t, db.addressCoinBalance["a2"]["c1"], 3)
	assert.Equal(t, db.addressCoinBalance["a1"]["c2"], 1)
	assert.Equal(t, len(db.coinAddressBalance["c1"]), 1)
	assert.Equal(t, len(db.addressCoinBalance["a1"]), 1)
	assert.Equal(t, db.coins["c1"].HolderCount, 1)
	assert.Equal(t, db.coins["c2"].HolderCount, 1)
}

func TestEncodeMapStringInt(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}

	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(m); err != nil {
		t.Fatal(err)
	}

	m2 := map[string]int{}
	if err := gob.NewDecoder(bytes.NewReader(data.Bytes())).Decode(&m2); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, m, m2)
}

func TestEncodeMapStringUnspentCoin(t *testing.T) {
	m := map[string]*types.UnspentCoin{
		"a": {
			CoinId: "c1",
			Amount: 1,
		},
		"b": {
			CoinId: "c2",
			Amount: 2,
		},
	}

	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(m); err != nil {
		t.Fatal(err)
	}

	m2 := map[string]*types.UnspentCoin{}
	if err := gob.NewDecoder(bytes.NewReader(data.Bytes())).Decode(&m2); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, m, m2)
}

func TestMemDbSaveLoadCoinInfo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		os.RemoveAll("./memdb-test-coins/")
	}()

	db := NewMemDb("./memdb-test-coins/", "testnet", false, logger)
	db.CoinInfoBatchUpdate(map[string]*types.CoinInfo{
		"c1": {
			Id:          "c1",
			TotalSupply: 1,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
			TxCount:     1,
			HolderCount: 1,
			CreatedAt:   2,
		},
		"c2": {
			Id:          "c2",
			TotalSupply: 1,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
			TxCount:     1,
			HolderCount: 1,
			CreatedAt:   2,
		},
	})
	db.IndexedHeightUpdate(1)
	db.Close()

	db2 := NewMemDb("./memdb-test-coins/", "testnet", false, logger)
	db2.persistDb.Close()

	assert.Equal(t, db.coins, db2.coins)
}

func TestMemDbSaveLoadUtxo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		os.RemoveAll("./memdb-test-utxos/")
	}()

	db := NewMemDb("./memdb-test-utxos/", "testnet", false, logger)
	db.UtxoBatchUpdate(map[string]*types.UnspentCoin{
		"u1": {
			CoinId: "c1",
			Owner:  "a1",
			Amount: 1,
			Utxo:   "u1",
		},
		"u2": {
			CoinId: "c2",
			Owner:  "a2",
			Amount: 2,
			Utxo:   "u2",
		},
	})
	db.IndexedHeightUpdate(1)
	db.Close()

	db2 := NewMemDb("./memdb-test-utxos/", "testnet", false, logger)
	db2.persistDb.Close()

	assert.Equal(t, db.utxoCoin, db2.utxoCoin)
	assert.Equal(t, db.addressUtxoCoin, db2.addressUtxoCoin)
}

func TestMemDbSaveLoadBalance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		os.RemoveAll("./memdb-test-balances/")
	}()

	db := NewMemDb("./memdb-test-balances/", "testnet", false, logger)
	db.CoinInfoBatchUpdate(map[string]*types.CoinInfo{
		"c1": {
			Id:          "c1",
			TotalSupply: 1,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
			TxCount:     1,
			HolderCount: 1,
			CreatedAt:   2,
		},
		"c2": {
			Id:          "c2",
			TotalSupply: 1,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
			TxCount:     1,
			HolderCount: 1,
			CreatedAt:   2,
		},
	})
	db.BalanceBatchUpdate(map[string]map[string]int{
		"c1": {
			"a1": 1,
			"a2": 2,
		},
		"c2": {
			"a1": 1,
		},
	})
	db.IndexedHeightUpdate(1)
	db.Close()

	db2 := NewMemDb("./memdb-test-balances/", "testnet", false, logger)
	db2.persistDb.Close()

	assert.Equal(t, db.coinAddressBalance, db2.coinAddressBalance)
	assert.Equal(t, db.addressCoinBalance, db2.addressCoinBalance)
}
