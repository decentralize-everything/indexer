package store

import (
	"github.com/decentralize-everything/indexer/types"
)

type Database interface {
	GetCoinInfos() ([]*types.CoinInfo, error)
	GetCoinInfoById(id string) (*types.CoinInfo, error)
	GetCoinsInUtxos(utxos []string) ([]*types.UnspentCoin, error)
	GetBalancesByAddress(address string) (map[string]int, error)
	GetCoinsByAddress(address string) ([]*types.UnspentCoin, error)
	CoinInfoBatchUpdate(updates map[string]*types.CoinInfo) error
	BalanceBatchUpdate(coinAddressBalances map[string]map[string]int) error
	UtxoBatchUpdate(updates map[string]*types.UnspentCoin) error
}
