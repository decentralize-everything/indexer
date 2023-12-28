package mempool

import (
	"github.com/decentralize-everything/indexer/extract"
)

type Vin struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

var _ extract.Vin = (*Vin)(nil)

func (v *Vin) GetTxid() string {
	return v.Txid
}

func (v *Vin) GetVout() int {
	return v.Vout
}

type Vout struct {
	Value   float64 `json:"value"`
	Address string  `json:"scriptpubkey_address"`
	Asm     string  `json:"scriptpubkey_asm"`
}

var _ extract.Vout = (*Vout)(nil)

func (v *Vout) GetValue() float64 {
	return v.Value
}

func (v *Vout) GetAddress() string {
	return v.Address
}

func (v *Vout) GetAsm() string {
	return v.Asm
}

type Transaction struct {
	Txid string `json:"txid"`
	Vin  []Vin  `json:"vin"`
	Vout []Vout `json:"vout"`
}

var _ extract.Transaction = (*Transaction)(nil)

func (t *Transaction) GetTxid() string {
	return t.Txid
}

func (t *Transaction) GetVin() []extract.Vin {
	vins := make([]extract.Vin, len(t.Vin))
	for i := range t.Vin {
		vins[i] = &t.Vin[i]
	}
	return vins
}

func (t *Transaction) GetVout() []extract.Vout {
	vouts := make([]extract.Vout, len(t.Vout))
	for i := range t.Vout {
		vouts[i] = &t.Vout[i]
	}
	return vouts
}

type Block struct {
	Hash   string
	Time   int
	Height int
	Tx     []Transaction
}

var _ extract.Block = (*Block)(nil)

func (b *Block) GetHash() string {
	return b.Hash
}

func (b *Block) GetTime() int {
	return b.Time
}

func (b *Block) GetHeight() int {
	return b.Height
}

func (b *Block) GetTxs() []extract.Transaction {
	txs := make([]extract.Transaction, len(b.Tx))
	for i := range b.Tx {
		txs[i] = &b.Tx[i]
	}
	return txs
}
