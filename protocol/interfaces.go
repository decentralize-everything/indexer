package protocol

import (
	"github.com/decentralize-everything/indexer/extract"
	"github.com/decentralize-everything/indexer/types"
)

type Parser interface {
	Parse(tx extract.Transaction) ([]*types.NewCoinEvent, []*types.BalanceChangeEvent, error)
}
