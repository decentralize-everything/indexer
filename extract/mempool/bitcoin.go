package mempool

import (
	"encoding/json"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/vincentdebug/go-ord-tx/pkg/btcapi/mempool"
)

type BitcoinClient struct {
	client *mempool.MempoolClient
}

func NewBitcoinClient(params *chaincfg.Params) *BitcoinClient {
	return &BitcoinClient{
		client: mempool.NewClient(params),
	}
}

func (c *BitcoinClient) GetLatestBlockHeight() (int, error) {
	panic("implement me")
}

func (c *BitcoinClient) GetBlockHash(blockHeight int) (string, error) {
	return c.client.GetBlockHash(blockHeight)
}

func (c *BitcoinClient) GetBlock(blockHash string) (*Block, error) {
	transactions, err := c.client.GetTransactions(blockHash)
	if err != nil {
		return nil, err
	}

	var txs []Transaction
	if err := json.Unmarshal([]byte(transactions), &txs); err != nil {
		return nil, err
	}

	var j []map[string]interface{}
	if err := json.Unmarshal([]byte(transactions), &j); err != nil {
		return nil, err
	}

	return &Block{
		Hash:   j[0]["status"].(map[string]interface{})["block_hash"].(string),
		Time:   int(j[0]["status"].(map[string]interface{})["block_time"].(float64)),
		Height: int(j[0]["status"].(map[string]interface{})["block_height"].(float64)),
		Tx:     txs,
	}, nil
}
