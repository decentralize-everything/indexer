package getblock

import (
	"context"

	"github.com/ybbus/jsonrpc/v3"
)

type BitcoinClient struct {
	client jsonrpc.RPCClient
}

func NewBitcoinClient(url string) *BitcoinClient {
	return &BitcoinClient{
		client: jsonrpc.NewClient(url),
	}
}

func (c *BitcoinClient) GetLatestBlockHeight() (int, error) {
	var result int
	err := c.client.CallFor(context.Background(), &result, "getblockcount")
	return result, err
}

func (c *BitcoinClient) GetBlockHash(blockHeight int) (string, error) {
	var result string
	err := c.client.CallFor(context.Background(), &result, "getblockhash", blockHeight)
	return result, err
}

func (c *BitcoinClient) GetBlock(blockHash string) (*Block, error) {
	// func (c *BitcoinClient) GetBlock(blockHash string) (map[string]interface{}, error) {
	var result Block
	// var result map[string]interface{}
	err := c.client.CallFor(context.Background(), &result, "getblock", blockHash, 2)
	return &result, err
}
