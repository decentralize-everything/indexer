package types

type CoinInfo struct {
	Id          string
	TotalSupply int
	Args        map[string]interface{}
	TxCount     int
	HolderCount int
	CreatedAt   int // Created at which block height.
}

type UnspentCoin struct {
	CoinId string
	Owner  string
	Amount int
	Utxo   string
}
