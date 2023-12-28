package store

import (
	"testing"

	"github.com/decentralize-everything/indexer/types"
	"github.com/stretchr/testify/assert"
)

func TestMemDbBalanceBatchUpdate(t *testing.T) {
	db := NewMemDb(false)
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
