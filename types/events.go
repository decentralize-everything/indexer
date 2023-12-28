package types

type NewCoinEvent struct {
	ChainId  string                 `json:"chain_id"`
	Protocol string                 `json:"protocol"`
	CoinId   string                 `json:"coin_id"`
	Args     map[string]interface{} `json:"args"`
}

type BalanceChangeEvent struct {
	ChainId  string `json:"chain_id"`
	Protocol string `json:"protocol"`
	CoinId   string `json:"coin_id"`
	Address  string `json:"address"`
	Delta    int    `json:"delta"`
	Utxo     string `json:"utxo"`
	IsMint   bool   `json:"is_mint"`
}
