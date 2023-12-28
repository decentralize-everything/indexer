package store

import (
	"sync"

	"github.com/decentralize-everything/indexer/types"
)

type MemDb struct {
	mutex              sync.RWMutex
	coins              map[string]*types.CoinInfo
	utxoCoin           map[string]*types.UnspentCoin
	addressCoin        map[string]map[string]*types.UnspentCoin
	addressCoinBalance map[string]map[string]int
	coinAddressBalance map[string]map[string]int
}

var _ Database = (*MemDb)(nil)

func NewMemDb(debug bool) *MemDb {
	db := &MemDb{
		coins:              make(map[string]*types.CoinInfo),
		utxoCoin:           make(map[string]*types.UnspentCoin),
		addressCoin:        make(map[string]map[string]*types.UnspentCoin),
		addressCoinBalance: make(map[string]map[string]int),
		coinAddressBalance: make(map[string]map[string]int),
	}

	if debug {
		db.fillTestData()
	}
	return db
}

func (m *MemDb) GetCoinInfos() ([]*types.CoinInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*types.CoinInfo
	for _, ci := range m.coins {
		results = append(results, ci)
	}
	return results, nil
}

func (m *MemDb) GetCoinInfoById(id string) (*types.CoinInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if ci, ok := m.coins[id]; ok {
		return ci, nil
	}
	return nil, nil
}

func (m *MemDb) GetCoinsInUtxos(utxoCoin []string) ([]*types.UnspentCoin, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*types.UnspentCoin
	for _, utxo := range utxoCoin {
		if uc, ok := m.utxoCoin[utxo]; ok {
			results = append(results, uc)
		}
	}
	return results, nil
}

func (m *MemDb) GetBalancesByAddress(address string) (map[string]int, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if balances, ok := m.addressCoinBalance[address]; ok {
		return balances, nil
	}
	return nil, nil
}

func (m *MemDb) GetCoinsByAddress(address string) ([]*types.UnspentCoin, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if coins, ok := m.addressCoin[address]; ok {
		var results []*types.UnspentCoin
		for _, coin := range coins {
			results = append(results, coin)
		}
		return results, nil
	}
	return nil, nil
}

func (m *MemDb) CoinInfoBatchUpdate(updates map[string]*types.CoinInfo) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, ci := range updates {
		m.coins[id] = ci
	}
	return nil
}

func (m *MemDb) BalanceBatchUpdate(coinAddressBalances map[string]map[string]int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for coin, balances := range coinAddressBalances {
		if _, ok := m.coinAddressBalance[coin]; !ok {
			m.coinAddressBalance[coin] = make(map[string]int)
		}
		for address, balance := range balances {
			m.coinAddressBalance[coin][address] += balance
			if m.coinAddressBalance[coin][address] == 0 {
				delete(m.coinAddressBalance[coin], address)
				delete(m.addressCoinBalance[address], coin)
			} else {
				if _, ok := m.addressCoinBalance[address]; !ok {
					m.addressCoinBalance[address] = make(map[string]int)
				}
				m.addressCoinBalance[address][coin] += balance
			}
		}

		// Update coin info.
		if _, ok := m.coins[coin]; ok {
			m.coins[coin].HolderCount = len(m.coinAddressBalance[coin])
		} else {
			panic("unexpected error: coin info didn't exist")
		}
	}
	return nil
}

func (m *MemDb) UtxoBatchUpdate(updates map[string]*types.UnspentCoin) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for utxo, uc := range updates {
		if uc == nil {
			delete(m.utxoCoin, utxo)
			delete(m.addressCoin[uc.Owner], uc.Utxo)
		} else {
			m.utxoCoin[utxo] = uc
			if _, ok := m.addressCoin[uc.Owner]; !ok {
				m.addressCoin[uc.Owner] = make(map[string]*types.UnspentCoin)
			}
			m.addressCoin[uc.Owner][uc.Utxo] = uc
		}
	}
	return nil
}

func (m *MemDb) fillTestData() {
	m.coins["TESTCA"] = &types.CoinInfo{
		Id:          "TESTCA",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     1,
		HolderCount: 100,
		CreatedAt:   800005,
	}
	m.coins["TESTCB"] = &types.CoinInfo{
		Id:          "TESTCB",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     2,
		HolderCount: 99,
		CreatedAt:   800006,
	}
	m.coins["TESTCC"] = &types.CoinInfo{
		Id:          "TESTCC",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     3,
		HolderCount: 98,
		CreatedAt:   800007,
	}
	m.coins["TESTCD"] = &types.CoinInfo{
		Id:          "TESTCD",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     4,
		HolderCount: 97,
		CreatedAt:   800008,
	}
	m.coins["TESTCE"] = &types.CoinInfo{
		Id:          "TESTCE",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     5,
		HolderCount: 96,
		CreatedAt:   800009,
	}
	m.coins["TESTCF"] = &types.CoinInfo{
		Id:          "TESTCF",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     6,
		HolderCount: 95,
		CreatedAt:   800010,
	}
	m.coins["TESTCG"] = &types.CoinInfo{
		Id:          "TESTCG",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     7,
		HolderCount: 94,
		CreatedAt:   800011,
	}
	m.coins["TESTCH"] = &types.CoinInfo{
		Id:          "TESTCH",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     8,
		HolderCount: 93,
		CreatedAt:   800012,
	}
	m.coins["TESTCI"] = &types.CoinInfo{
		Id:          "TESTCI",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     9,
		HolderCount: 92,
		CreatedAt:   800001,
	}
	m.coins["TESTCJ"] = &types.CoinInfo{
		Id:          "TESTCJ",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     10,
		HolderCount: 91,
		CreatedAt:   800002,
	}
	m.coins["TESTCK"] = &types.CoinInfo{
		Id:          "TESTCK",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     11,
		HolderCount: 90,
		CreatedAt:   800003,
	}
	m.coins["TESTCL"] = &types.CoinInfo{
		Id:          "TESTCL",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":  uint64(100),
			"sats": uint64(10000),
		},
		TxCount:     12,
		HolderCount: 89,
		CreatedAt:   800004,
	}

	m.coinAddressBalance["TESTCA"] = map[string]int{
		"addr1": 1,
		"addr2": 2,
	}
	m.coinAddressBalance["TESTCB"] = map[string]int{
		"addr1": 3,
		"addr2": 4,
	}
	m.addressCoinBalance["addr1"] = map[string]int{
		"TESTCA": 1,
		"TESTCB": 3,
	}
	m.addressCoinBalance["addr2"] = map[string]int{
		"TESTCA": 2,
		"TESTCB": 4,
	}

	m.utxoCoin["1111:0"] = &types.UnspentCoin{
		CoinId: "TESTCA",
		Owner:  "addr1",
		Amount: 1,
		Utxo:   "1111:0",
	}
	m.utxoCoin["1112:0"] = &types.UnspentCoin{
		CoinId: "TESTCA",
		Owner:  "addr2",
		Amount: 2,
		Utxo:   "1112:0",
	}
	m.utxoCoin["1113:0"] = &types.UnspentCoin{
		CoinId: "TESTCB",
		Owner:  "addr1",
		Amount: 3,
		Utxo:   "1113:0",
	}
	m.utxoCoin["1114:0"] = &types.UnspentCoin{
		CoinId: "TESTCB",
		Owner:  "addr2",
		Amount: 4,
		Utxo:   "1114:0",
	}

	m.addressCoin["addr1"] = map[string]*types.UnspentCoin{
		"1111:0": m.utxoCoin["1111:0"],
		"1113:0": m.utxoCoin["1113:0"],
	}
	m.addressCoin["addr2"] = map[string]*types.UnspentCoin{
		"1112:0": m.utxoCoin["1112:0"],
		"1114:0": m.utxoCoin["1114:0"],
	}
}
