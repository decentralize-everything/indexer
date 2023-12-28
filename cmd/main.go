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
	Height     int    `help:"Starting block height" default:"823122"`
	Network    string `help:"Network working on, support 'testnet' or 'mainnet'" default:"mainnet"`
	Debug      bool   `help:"Enable debug mode"`
	DbFilePath string `help:"Database file path, disable persistent store by using --db-file-path=\"\"" default:"./indexer.db"`
}

func main() {
	kong.Parse(
		&cli,
		kong.Name("indexer"),
		kong.Description("Indexer for Carve Coin protocol"),
		kong.UsageOnError(),
	)

	var params *chaincfg.Params
	switch cli.Network {
	case "testnet":
		params = &chaincfg.TestNet3Params
	case "mainnet":
		params = &chaincfg.MainNetParams
	default:
		panic("invalid network")
	}

	logger, _ := zap.NewDevelopment()
	db := store.NewMemDb(cli.DbFilePath, cli.Network, cli.Debug, logger.Named("store"))
	btcClient := mempool.NewBitcoinClient(params)
	btcTransformer := transform.NewBitcoinTransformer(db, logger.Named("transform"))
	updater := load.NewDbUpdater(db, logger.Named("load"))

	height, network, err := db.GetStatus()
	if err != nil {
		panic(err)
	}
	if height != 0 {
		height++
		logger.Warn("Overwrite command line parameter", zap.Int("height", height), zap.String("network", network))
	} else {
		height = cli.Height
	}

	// Start http service.
	router := api.SetupRouter(db)
	go router.Run(":8080")

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
