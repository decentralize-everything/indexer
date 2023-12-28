package transform

import (
	"github.com/decentralize-everything/indexer/extract"
	"github.com/decentralize-everything/indexer/protocol"
	"github.com/decentralize-everything/indexer/store"
	"github.com/decentralize-everything/indexer/types"
	"go.uber.org/zap"
)

type BitcoinTransformer struct {
	protocols []protocol.Parser
	db        store.Database
	logger    *zap.Logger
}

func NewBitcoinTransformer(db store.Database, logger *zap.Logger) *BitcoinTransformer {
	return &BitcoinTransformer{
		protocols: []protocol.Parser{
			protocol.NewCarvProtocol(db, logger),
		},
		db:     db,
		logger: logger,
	}
}

func (t *BitcoinTransformer) Transform(block extract.Block) (*types.BatchUpdate, error) {
	batchUpdate := &types.BatchUpdate{
		Block: block,
	}

	for _, tx := range block.GetTxs() {
		for _, protocol := range t.protocols {
			newCoinEvents, balanceChangeEvents, err := protocol.Parse(tx)
			if err != nil {
				t.logger.Warn("protocol.Parse", zap.Error(err))
			}

			if len(newCoinEvents) > 0 || len(balanceChangeEvents) > 0 {
				batchUpdate.TxUpdates = append(batchUpdate.TxUpdates, &types.TxUpdate{
					Txid:                tx.GetTxid(),
					NewCoinEvents:       newCoinEvents,
					BalanceChangeEvents: balanceChangeEvents,
				})
			}
		}
	}
	return batchUpdate, nil
}
