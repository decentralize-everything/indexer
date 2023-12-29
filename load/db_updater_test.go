package load

import (
	"testing"

	"github.com/decentralize-everything/indexer/extract/mempool"
	"github.com/decentralize-everything/indexer/store"
	"github.com/decentralize-everything/indexer/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func setup(t *testing.T) (*store.MockDatabase, *DbUpdater, *observer.ObservedLogs) {
	observedZapCore, observedLogs := observer.New(zap.DebugLevel)
	logger := zap.New(observedZapCore)
	ctrl := gomock.NewController(t)
	mockDb := store.NewMockDatabase(ctrl)
	return mockDb, NewDbUpdater(mockDb, logger), observedLogs
}

func TestDeployDuplicateCoinOnSameBlockError(t *testing.T) {
	mockDb, updater, observedLogs := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(nil, nil)
	mockDb.EXPECT().CoinInfoBatchUpdate(gomock.Any())
	mockDb.EXPECT().IndexedHeightUpdate(1)

	updater.Update(&types.BatchUpdate{
		Block: &mempool.Block{
			Height: 1,
		},
		TxUpdates: []*types.TxUpdate{
			{
				Txid: "1234",
				NewCoinEvents: []*types.NewCoinEvent{
					{
						CoinId: "CARV",
					},
				},
			},
			{
				Txid: "5678",
				NewCoinEvents: []*types.NewCoinEvent{
					{
						CoinId: "CARV",
					},
				},
			},
		},
	})

	assert.Equal(t, 1, observedLogs.Len())
	allLogs := observedLogs.All()
	assert.Equal(t, "duplicated coin deployment transaction on same block", allLogs[0].Message)
	assert.ElementsMatch(t, []zap.Field{
		{Key: "id", Type: zapcore.StringType, String: "CARV"},
		{Key: "tx", Type: zapcore.StringType, String: "5678"},
	}, allLogs[0].Context)
}

func TestMintNonExistCoinError(t *testing.T) {
	mockDb, updater, observedLogs := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(nil, nil)
	mockDb.EXPECT().IndexedHeightUpdate(1)

	updater.Update(&types.BatchUpdate{
		Block: &mempool.Block{
			Height: 1,
		},
		TxUpdates: []*types.TxUpdate{
			{
				Txid: "1234",
				BalanceChangeEvents: []*types.BalanceChangeEvent{
					{
						CoinId: "CARV",
						IsMint: true,
					},
				},
			},
		},
	})

	assert.Equal(t, 1, observedLogs.Len())
	allLogs := observedLogs.All()
	assert.Equal(t, "mint or transfer on a non-exist coin", allLogs[0].Message)
	assert.ElementsMatch(t, []zap.Field{
		{Key: "id", Type: zapcore.StringType, String: "CARV"},
		{Key: "tx", Type: zapcore.StringType, String: "1234"},
	}, allLogs[0].Context)
}

func TestMintExceedMaxSupplyError(t *testing.T) {
	mockDb, updater, observedLogs := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 100,
		Args: map[string]interface{}{
			"max": uint64(100),
		},
	}, nil)
	mockDb.EXPECT().CoinInfoBatchUpdate(map[string]*types.CoinInfo{
		"CARV": {
			Id:          "CARV",
			TotalSupply: 100,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
		},
	})
	mockDb.EXPECT().IndexedHeightUpdate(1)

	updater.Update(&types.BatchUpdate{
		Block: &mempool.Block{
			Height: 1,
		},
		TxUpdates: []*types.TxUpdate{
			{
				Txid: "1234",
				BalanceChangeEvents: []*types.BalanceChangeEvent{
					{
						CoinId: "CARV",
						IsMint: true,
						Delta:  1,
					},
				},
			},
		},
	})

	assert.Equal(t, 1, observedLogs.Len())
	allLogs := observedLogs.All()
	assert.Equal(t, "mint exceed max supply", allLogs[0].Message)
	assert.ElementsMatch(t, []zap.Field{
		{Key: "id", Type: zapcore.StringType, String: "CARV"},
		{Key: "tx", Type: zapcore.StringType, String: "1234"},
	}, allLogs[0].Context)
}

func TestDeploySuccess(t *testing.T) {
	mockDb, updater, _ := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(nil, nil)
	mockDb.EXPECT().CoinInfoBatchUpdate(map[string]*types.CoinInfo{
		"CARV": {
			Id:          "CARV",
			TotalSupply: 0,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
			TxCount:      1,
			CreatedAt:    1234567890,
			DeployTx:     "1234",
			DeployHeight: 1,
		},
	})
	mockDb.EXPECT().IndexedHeightUpdate(1)

	updater.Update(&types.BatchUpdate{
		Block: &mempool.Block{
			Height: 1,
			Time:   1234567890,
		},
		TxUpdates: []*types.TxUpdate{
			{
				Txid: "1234",
				NewCoinEvents: []*types.NewCoinEvent{
					{
						CoinId: "CARV",
						Args: map[string]interface{}{
							"max": uint64(100),
						},
					},
				},
			},
		},
	})
}

func TestMintSuccess(t *testing.T) {
	mockDb, updater, _ := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max": uint64(100),
		},
	}, nil)
	mockDb.EXPECT().CoinInfoBatchUpdate(map[string]*types.CoinInfo{
		"CARV": {
			Id:          "CARV",
			TotalSupply: 2,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
			TxCount: 1,
		},
	})
	mockDb.EXPECT().BalanceBatchUpdate(map[string]map[string]int{
		"CARV": {
			"5678": 1,
		},
	})
	mockDb.EXPECT().UtxoBatchUpdate(map[string]*types.UnspentCoin{
		"1234:0": {
			CoinId: "CARV",
			Owner:  "5678",
			Amount: 1,
			Utxo:   "1234:0",
		},
	})
	mockDb.EXPECT().IndexedHeightUpdate(1)

	updater.Update(&types.BatchUpdate{
		Block: &mempool.Block{
			Height: 1,
		},
		TxUpdates: []*types.TxUpdate{
			{
				Txid: "1234",
				BalanceChangeEvents: []*types.BalanceChangeEvent{
					{
						CoinId:  "CARV",
						Address: "5678",
						IsMint:  true,
						Delta:   1,
						Utxo:    "1234:0",
					},
				},
			},
		},
	})
}

func TestTransferSuccess(t *testing.T) {
	mockDb, updater, _ := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max": uint64(100),
		},
	}, nil)
	mockDb.EXPECT().CoinInfoBatchUpdate(map[string]*types.CoinInfo{
		"CARV": {
			Id:          "CARV",
			TotalSupply: 1,
			Args: map[string]interface{}{
				"max": uint64(100),
			},
			TxCount: 1,
		},
	})
	mockDb.EXPECT().BalanceBatchUpdate(map[string]map[string]int{
		"CARV": {
			"5678": -1,
			"1234": 1,
		},
	})
	mockDb.EXPECT().UtxoBatchUpdate(map[string]*types.UnspentCoin{
		"1234:0": {
			CoinId: "CARV",
			Owner:  "1234",
			Amount: 1,
			Utxo:   "1234:0",
		},
		"9abc:0": nil,
	})
	mockDb.EXPECT().IndexedHeightUpdate(1)

	updater.Update(&types.BatchUpdate{
		Block: &mempool.Block{
			Height: 1,
		},
		TxUpdates: []*types.TxUpdate{
			{
				Txid: "1234",
				BalanceChangeEvents: []*types.BalanceChangeEvent{
					{
						CoinId:  "CARV",
						Address: "5678",
						IsMint:  false,
						Delta:   -1,
						Utxo:    "9abc:0",
					},
					{
						CoinId:  "CARV",
						Address: "1234",
						IsMint:  false,
						Delta:   1,
						Utxo:    "1234:0",
					},
				},
			},
		},
	})
}
