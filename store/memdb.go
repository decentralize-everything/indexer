package store

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/decentralize-everything/indexer/types"
	"go.uber.org/zap"
)

/*
Data volume estimation:
- 1 coins < 1k
- 2 utxoCoin < 100k per coin = 100m
- 3 addressUtxoCoin < 100k addresses
- 4 addressCoinBalance < 100k addresses
- 5 coinAddressBalance < 1k coins
*/
type MemDb struct {
	mutex              sync.RWMutex
	network            string
	height             int
	coins              map[string]*types.CoinInfo
	utxoCoin           map[string]*types.UnspentCoin
	addressUtxoCoin    map[string]map[string]*types.UnspentCoin
	addressCoinBalance map[string]map[string]int
	coinAddressBalance map[string]map[string]int

	/*
		Data schema:
		- height: {"height" : {height}}
		- coins: {"coins/{coinId}" : {coinInfo}}
		- utxoCoin: {"utxos/{utxo}" : {unspentCoin}}
		- addressUtxoCoin: {"a-u-c/{address}" : {"{utxo}" : {unspentCoin}}}
		- addressCoinBalance: {"a-c-b/{address}" : {"{coinId}" : {balance}}}
		- coinAddressBalance: {"c-a-b/{coinId}" : {"{address}" : {balance}}}
	*/
	persistDb *BadgerDB
	logger    *zap.Logger
}

var (
	STATUS_KEY   = "status"
	COINS_PREFIX = "coins/"
	UTXOS_PREFIX = "utxos/"
	AUC_PREFIX   = "a-u-c/"
	ACB_PREFIX   = "a-c-b/"
	CAB_PREFIX   = "c-a-b/"
)

var _ Database = (*MemDb)(nil)

func NewMemDb(persistPath string, network string, debug bool, logger *zap.Logger) *MemDb {
	db := &MemDb{
		network:            network,
		coins:              make(map[string]*types.CoinInfo),
		utxoCoin:           make(map[string]*types.UnspentCoin),
		addressUtxoCoin:    make(map[string]map[string]*types.UnspentCoin),
		addressCoinBalance: make(map[string]map[string]int),
		coinAddressBalance: make(map[string]map[string]int),
		logger:             logger,
	}

	if len(persistPath) > 0 {
		db.persistDb = NewBadgerDB(persistPath)
		db.loadIntoMem()
	}

	if debug {
		db.fillTestData()
	}
	return db
}

func (m *MemDb) Close() error {
	if m.persistDb != nil {
		return m.persistDb.Close()
	}
	return nil
}

func (m *MemDb) GetStatus() (int, string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.height, m.network, nil
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

	if coins, ok := m.addressUtxoCoin[address]; ok {
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

	var keys []string
	var values [][]byte
	for id, ci := range updates {
		m.coins[id] = ci
		if m.persistDb != nil {
			keys = append(keys, COINS_PREFIX+id)
			values = append(values, ci.ToBytes())
		}
	}

	if m.persistDb == nil {
		return nil
	}

	if err := m.persistDb.BatchSet(keys, values); err != nil {
		return err
	}
	return nil
}

func (m *MemDb) BalanceBatchUpdate(coinAddressBalances map[string]map[string]int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	uniqueAddresses := make(map[string]bool)
	uniqueCoins := make(map[string]bool)
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

			if m.persistDb != nil {
				uniqueAddresses[address] = true
			}
		}

		// Update coin info.
		if _, ok := m.coins[coin]; ok {
			m.coins[coin].HolderCount = len(m.coinAddressBalance[coin])
		} else {
			panic("unexpected error: coin info didn't exist")
		}

		if m.persistDb != nil {
			uniqueCoins[coin] = true
		}
	}

	if m.persistDb == nil {
		return nil
	}

	var keys []string
	var values [][]byte
	for coin := range uniqueCoins {
		keys = append(keys, CAB_PREFIX+coin)
		var data bytes.Buffer
		if err := gob.NewEncoder(&data).Encode(m.coinAddressBalance[coin]); err != nil {
			return err
		}
		values = append(values, data.Bytes())
	}
	for address := range uniqueAddresses {
		keys = append(keys, ACB_PREFIX+address)
		var data bytes.Buffer
		if err := gob.NewEncoder(&data).Encode(m.addressCoinBalance[address]); err != nil {
			return err
		}
		values = append(values, data.Bytes())
	}

	if err := m.persistDb.BatchSet(keys, values); err != nil {
		return err
	}
	return nil
}

func (m *MemDb) UtxoBatchUpdate(updates map[string]*types.UnspentCoin) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var keys []string
	var values [][]byte
	uniqueAddresses := make(map[string]bool)
	for utxo, uc := range updates {
		old := m.utxoCoin[utxo]
		if uc == nil {
			delete(m.utxoCoin, utxo)
			delete(m.addressUtxoCoin[old.Owner], old.Utxo)
		} else {
			m.utxoCoin[utxo] = uc
			if _, ok := m.addressUtxoCoin[uc.Owner]; !ok {
				m.addressUtxoCoin[uc.Owner] = make(map[string]*types.UnspentCoin)
			}
			m.addressUtxoCoin[uc.Owner][uc.Utxo] = uc
		}

		if m.persistDb != nil {
			keys = append(keys, UTXOS_PREFIX+utxo)
			if uc == nil {
				values = append(values, nil)
				uniqueAddresses[old.Owner] = true
			} else {
				values = append(values, uc.ToBytes())
				uniqueAddresses[uc.Owner] = true
			}
		}
	}

	if m.persistDb == nil {
		return nil
	}

	// Collect unique addresses' updates.
	for address := range uniqueAddresses {
		keys = append(keys, AUC_PREFIX+address)
		var data bytes.Buffer
		if err := gob.NewEncoder(&data).Encode(m.addressUtxoCoin[address]); err != nil {
			return err
		}
		values = append(values, data.Bytes())
	}

	if err := m.persistDb.BatchSet(keys, values); err != nil {
		return err
	}
	return nil
}

func (m *MemDb) IndexedHeightUpdate(height int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.height = height

	if m.persistDb != nil {
		var keys []string
		var values [][]byte
		keys = append(keys, STATUS_KEY)
		var data bytes.Buffer
		if err := gob.NewEncoder(&data).Encode(map[string]interface{}{"height": m.height, "network": m.network}); err != nil {
			return err
		}
		values = append(values, data.Bytes())

		if err := m.persistDb.BatchSet(keys, values); err != nil {
			return err
		}
		m.persistDb.Sync()
	}
	return nil
}

func (m *MemDb) loadIntoMem() {
	m.logger.Info("loading data from disk into memory")
	start := time.Now()
	defer func() {
		m.logger.Info("loading data from disk into memory done", zap.Duration("duration", time.Since(start)))
	}()

	// Load height.
	v, err := m.persistDb.Get(STATUS_KEY)
	if err != nil || v == nil {
		m.height = 0
	} else {
		status := make(map[string]interface{})
		if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&status); err != nil {
			panic(fmt.Sprintf("failed to decode height from disk: %v", err))
		}
		m.height = status["height"].(int)
		m.network = status["network"].(string)
	}

	if m.height == 0 {
		return
	}

	// Load coins.
	_, values, err := m.persistDb.Query(COINS_PREFIX)
	if err != nil {
		panic(fmt.Sprintf("failed to load coins from disk: %v", err))
	}
	for i := range values {
		ci := &types.CoinInfo{}
		if err := ci.FromBytes(values[i]); err != nil {
			panic(fmt.Sprintf("failed to decode coin info from disk: %v", err))
		}
		m.coins[ci.Id] = ci
	}

	// Load utxoCoin.
	_, values, err = m.persistDb.Query(UTXOS_PREFIX)
	if err != nil {
		panic(fmt.Sprintf("failed to load utxoCoin from disk: %v", err))
	}
	for i := range values {
		uc := &types.UnspentCoin{}
		if err := uc.FromBytes(values[i]); err != nil {
			panic(fmt.Sprintf("failed to decode utxoCoin from disk: %v", err))
		}
		m.utxoCoin[uc.Utxo] = uc
	}

	// Load addressUtxoCoin.
	keys, values, err := m.persistDb.Query(AUC_PREFIX)
	if err != nil {
		panic(fmt.Sprintf("failed to load addressUtxoCoin from disk: %v", err))
	}
	for i := range values {
		address := keys[i][len(AUC_PREFIX):]
		utxoCoin := make(map[string]*types.UnspentCoin)
		if err := gob.NewDecoder(bytes.NewReader(values[i])).Decode(&utxoCoin); err != nil {
			panic(fmt.Sprintf("failed to decode addressUtxoCoin from disk: %v", err))
		}
		m.addressUtxoCoin[address] = utxoCoin
	}

	// Load addressCoinBalance.
	keys, values, err = m.persistDb.Query(ACB_PREFIX)
	if err != nil {
		panic(fmt.Sprintf("failed to load addressCoinBalance from disk: %v", err))
	}
	for i := range values {
		address := keys[i][len(ACB_PREFIX):]
		balance := make(map[string]int)
		if err := gob.NewDecoder(bytes.NewReader(values[i])).Decode(&balance); err != nil {
			panic(fmt.Sprintf("failed to decode addressCoinBalance from disk: %v", err))
		}
		m.addressCoinBalance[address] = balance
	}

	// Load coinAddressBalance.
	keys, values, err = m.persistDb.Query(CAB_PREFIX)
	if err != nil {
		panic(fmt.Sprintf("failed to load coinAddressBalance from disk: %v", err))
	}
	for i := range values {
		coin := keys[i][len(CAB_PREFIX):]
		balance := make(map[string]int)
		if err := gob.NewDecoder(bytes.NewReader(values[i])).Decode(&balance); err != nil {
			panic(fmt.Sprintf("failed to decode coinAddressBalance from disk: %v", err))
		}
		m.coinAddressBalance[coin] = balance
	}
}

func (m *MemDb) fillTestData() {
	m.coins["TESTCA"] = &types.CoinInfo{
		Id:          "TESTCA",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      1,
		HolderCount:  100,
		CreatedAt:    1703500000,
		DeployTx:     "1111",
		DeployHeight: 800005,
	}
	m.coins["TESTCB"] = &types.CoinInfo{
		Id:          "TESTCB",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      2,
		HolderCount:  99,
		CreatedAt:    1703600000,
		DeployTx:     "2222",
		DeployHeight: 800006,
	}
	m.coins["TESTCC"] = &types.CoinInfo{
		Id:          "TESTCC",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      3,
		HolderCount:  98,
		CreatedAt:    1703700000,
		DeployTx:     "3333",
		DeployHeight: 800007,
	}
	m.coins["TESTCD"] = &types.CoinInfo{
		Id:          "TESTCD",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      4,
		HolderCount:  97,
		CreatedAt:    1703800000,
		DeployTx:     "4444",
		DeployHeight: 800008,
	}
	m.coins["TESTCE"] = &types.CoinInfo{
		Id:          "TESTCE",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      5,
		HolderCount:  96,
		CreatedAt:    1703900000,
		DeployTx:     "5555",
		DeployHeight: 800009,
	}
	m.coins["TESTCF"] = &types.CoinInfo{
		Id:          "TESTCF",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      6,
		HolderCount:  95,
		CreatedAt:    1704000000,
		DeployTx:     "6666",
		DeployHeight: 800010,
	}
	m.coins["TESTCG"] = &types.CoinInfo{
		Id:          "TESTCG",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      7,
		HolderCount:  94,
		CreatedAt:    1704100000,
		DeployTx:     "7777",
		DeployHeight: 800011,
	}
	m.coins["TESTCH"] = &types.CoinInfo{
		Id:          "TESTCH",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      8,
		HolderCount:  93,
		CreatedAt:    1704200000,
		DeployTx:     "8888",
		DeployHeight: 800012,
	}
	m.coins["TESTCI"] = &types.CoinInfo{
		Id:          "TESTCI",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      9,
		HolderCount:  92,
		CreatedAt:    1703100000,
		DeployTx:     "9999",
		DeployHeight: 800001,
	}
	m.coins["TESTCJ"] = &types.CoinInfo{
		Id:          "TESTCJ",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      10,
		HolderCount:  91,
		CreatedAt:    1703200000,
		DeployTx:     "aaaa",
		DeployHeight: 800002,
	}
	m.coins["TESTCK"] = &types.CoinInfo{
		Id:          "TESTCK",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      11,
		HolderCount:  90,
		CreatedAt:    1703300000,
		DeployTx:     "bbbb",
		DeployHeight: 800003,
	}
	m.coins["TESTCL"] = &types.CoinInfo{
		Id:          "TESTCL",
		TotalSupply: 5,
		Args: map[string]interface{}{
			"max":   uint64(100),
			"sats":  uint64(10000),
			"limit": 1,
		},
		TxCount:      12,
		HolderCount:  89,
		CreatedAt:    1703400000,
		DeployTx:     "cccc",
		DeployHeight: 800004,
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

	m.addressUtxoCoin["addr1"] = map[string]*types.UnspentCoin{
		"1111:0": m.utxoCoin["1111:0"],
		"1113:0": m.utxoCoin["1113:0"],
	}
	m.addressUtxoCoin["addr2"] = map[string]*types.UnspentCoin{
		"1112:0": m.utxoCoin["1112:0"],
		"1114:0": m.utxoCoin["1114:0"],
	}
}
