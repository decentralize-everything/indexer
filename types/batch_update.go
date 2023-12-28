package types

import "github.com/decentralize-everything/indexer/extract"

type TxUpdate struct {
	Txid                string
	NewCoinEvents       []*NewCoinEvent
	BalanceChangeEvents []*BalanceChangeEvent
}

type BatchUpdate struct {
	Block     extract.Block
	TxUpdates []*TxUpdate
}
