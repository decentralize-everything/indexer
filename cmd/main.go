package main

import (
	"time"

	"github.com/alecthomas/kong"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decentralize-everything/indexer/api"
	"github.com/decentralize-everything/indexer/extract/mempool"
	"github.com/decentralize-everything/indexer/load"
	"github.com/decentralize-everything/indexer/store"
	"github.com/decentralize-everything/indexer/transform"
	"go.uber.org/zap"
)

var cli struct {
	Height int  `help:"Starting block height" default:"823122"`
	Debug  bool `help:"Enable debug mode"`
}

func main() {
	kong.Parse(
		&cli,
		kong.Name("indexer"),
		kong.Description("Indexer for Carve Coin protocol"),
		kong.UsageOnError(),
	)

	logger, _ := zap.NewDevelopment()
	db := store.NewMemDb(cli.Debug)
	btcClient := mempool.NewBitcoinClient(&chaincfg.MainNetParams)
	btcTransformer := transform.NewBitcoinTransformer(db, logger.Named("transform"))
	updater := load.NewDbUpdater(db, logger.Named("load"))

	// Start http service.
	router := api.SetupRouter(db)
	go router.Run(":8080")

	height := cli.Height
	for {
		blockHash, err := btcClient.GetBlockHash(height)
		if err != nil {
			logger.Warn("btcClient.GetBlockHash", zap.Error(err))
			continue
		}

		block, err := btcClient.GetBlock(blockHash)
		if err != nil {
			// logger.Warn("btcClient.GetBlock", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		batchUpdate, err := btcTransformer.Transform(block)
		if err != nil {
			logger.Warn("btcTransformer.Transform", zap.Error(err))
			continue
		}

		if err := updater.Update(batchUpdate); err != nil {
			logger.Warn("updater.Update", zap.Error(err))
			continue
		}

		logger.Debug("Block processed", zap.Int("height", height))
		height++
	}
}
