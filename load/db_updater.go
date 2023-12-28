package load

import (
	"github.com/decentralize-everything/indexer/store"
	"github.com/decentralize-everything/indexer/types"
	"go.uber.org/zap"
)

type DbUpdater struct {
	db     store.Database
	logger *zap.Logger
}

func NewDbUpdater(db store.Database, logger *zap.Logger) *DbUpdater {
	return &DbUpdater{
		db:     db,
		logger: logger,
	}
}

func (u *DbUpdater) Update(batch *types.BatchUpdate) error {
	// Merge updates for batch operations.
	coinAddressBalanceUpdates := make(map[string]map[string]int)
	coinInfoUpdates := make(map[string]*types.CoinInfo)
	utxoUpdates := make(map[string]*types.UnspentCoin)

OUTER:
	for _, txUpdate := range batch.TxUpdates {
		for _, event := range txUpdate.NewCoinEvents {
			if _, ok := coinInfoUpdates[event.CoinId]; ok {
				u.logger.Info("duplicated coin deployment transaction on same block", zap.String("id", event.CoinId), zap.String("tx", txUpdate.Txid))
				continue OUTER // If this is a invalid deployment, than the whole transaction should be skipped.
			}

			if ci, err := u.db.GetCoinInfoById(event.CoinId); err == nil && ci == nil {
				coinInfoUpdates[event.CoinId] = &types.CoinInfo{
					Id:          event.CoinId,
					TotalSupply: 0,
					Args:        event.Args,
					TxCount:     0, // Set to zero here, because it will be incremented by mint later.
					CreatedAt:   batch.Block.GetHeight(),
				}
			} else {
				panic("unexpected error: duplicated coin deployment should be identified by transformer")
			}
		}

		for i, event := range txUpdate.BalanceChangeEvents {
			var ci *types.CoinInfo
			var err error
			ci, ok := coinInfoUpdates[event.CoinId]
			if !ok {
				ci, err = u.db.GetCoinInfoById(event.CoinId)
				if err != nil || ci == nil {
					u.logger.Info("mint or transfer on a non-exist coin", zap.String("id", event.CoinId), zap.String("tx", txUpdate.Txid))
					continue OUTER
				}
				coinInfoUpdates[event.CoinId] = ci
			}

			// Check total supply.
			if event.IsMint {
				if uint64(ci.TotalSupply+event.Delta) > ci.Args["max"].(uint64) {
					u.logger.Info("mint exceed max supply", zap.String("id", event.CoinId), zap.String("tx", txUpdate.Txid))
					continue OUTER
				}
				ci.TotalSupply += event.Delta
			}

			if _, ok := coinAddressBalanceUpdates[event.CoinId]; !ok {
				coinAddressBalanceUpdates[event.CoinId] = make(map[string]int)
			}
			coinAddressBalanceUpdates[event.CoinId][event.Address] += event.Delta

			if event.Delta > 0 {
				utxoUpdates[event.Utxo] = &types.UnspentCoin{
					CoinId: event.CoinId,
					Owner:  event.Address,
					Amount: event.Delta,
					Utxo:   event.Utxo,
				}
			} else { // Mark as delete.
				utxoUpdates[event.Utxo] = nil
			}

			// Currently, we suppose one transaction contains one type of coin operation.
			if i == 0 {
				ci.TxCount++
			}
		}
	}

	if len(coinInfoUpdates) > 0 {
		u.db.CoinInfoBatchUpdate(coinInfoUpdates)
	}
	if len(coinAddressBalanceUpdates) > 0 {
		u.db.BalanceBatchUpdate(coinAddressBalanceUpdates)
	}
	if len(utxoUpdates) > 0 {
		u.db.UtxoBatchUpdate(utxoUpdates)
	}
	return nil
}
