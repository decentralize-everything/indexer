package protocol

import (
	"reflect"
	"testing"

	"github.com/decentralize-everything/indexer/extract/mempool"
	"github.com/decentralize-everything/indexer/store"
	"github.com/decentralize-everything/indexer/types"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setup(t *testing.T) (*store.MockDatabase, *CarvProtocol) {
	logger, _ := zap.NewDevelopment()
	ctrl := gomock.NewController(t)
	mockDb := store.NewMockDatabase(ctrl)
	return mockDb, NewCarvProtocol(mockDb, logger)
}

func TestMetadataTooShortError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES",
				},
			},
		},
	)

	if err == nil || err.Error() != "metadata is too short: OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMetadataFormatError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_1 00 00",
				},
			},
		},
	)

	if err == nil || err.Error() != "invalid metadata format: OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_1 00 00" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMetadataLengthDecodeError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_X 00",
				},
			},
		},
	)

	if err == nil || err.Error() != "error parsing metadata length: OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_X 00" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMetadataLengthMismatchError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_1 0000",
				},
			},
		},
	)

	if err == nil || err.Error() != "metadata length mismatch: OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_1 0000" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMetadataDecodeError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_1 XX",
				},
			},
		},
	)

	if err == nil || err.Error() != "failed to decode metadata into bytes: OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_1 XX" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvalidArgumentNum(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_2 0000",
				},
			},
		},
	)

	if err == nil || err.Error() != "invalid Carv protocol metadata: OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_2 0000" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvalidCoinIdLen(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_9 8a80eb9e2501cd1001",
				},
			},
		},
	)

	if err == nil || err.Error() != "invalid arguments for deployment, id = INVALID, max = 1, sats = 10000, limit = 1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvalidMaxSupply(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_7 82a40500cd1001",
				},
			},
		},
	)

	if err == nil || err.Error() != "invalid arguments for deployment, id = CARV, max = 0, sats = 10000, limit = 1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvalidMinSats(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_7 82a40501a60801",
				},
			},
		},
	)

	if err == nil || err.Error() != "invalid arguments for deployment, id = CARV, max = 1, sats = 5000, limit = 1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLockedBtcTooHigh(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_11 82a4058980dd40bc834101",
				},
			},
		},
	)

	if err == nil || err.Error() != "locked BTC out of range, max = 21000000, sats = 1000001, lockedBtc = 21000021000000" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLockedBtcOverflow(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_15 82a4058cefacd5b9ba8eff00cd1001",
				},
			},
		},
	)

	if err == nil || err.Error() != "locked BTC out of range, max = 1000000000000000000, sats = 10000, lockedBtc = 1864712049423024128" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeployUtxoIsNot1stError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{
					Address: "1234",
					Value:   10000,
				},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_10 82a4058980dd40cd1001",
				},
			},
		},
	)

	if err == nil || err.Error() != "metadata for deployment placed at the 2-th UTXO, should be the first" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeployCoinAlreadyExists(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id: "CARV",
	}, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_10 82a4058980dd40cd1001",
				},
			},
		},
	)

	if err == nil || err.Error() != "coin ID already taken: CARV" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeploySuccess(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(nil, nil)

	newCoinEvents, balanceExchangeEvents, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_10 82a4058980dd40cd1001",
				},
			},
		},
	)

	if err != nil || len(newCoinEvents) != 1 || len(balanceExchangeEvents) != 0 {
		t.Fatal("should have found one new coin event")
	}

	expected := &types.NewCoinEvent{
		ChainId:  "bitcoin",
		Protocol: "carv",
		CoinId:   "CARV",
		Args: map[string]interface{}{
			"max":   uint64(21000000),
			"sats":  uint64(10000),
			"limit": uint64(1),
		},
	}
	if !reflect.DeepEqual(newCoinEvents[0], expected) {
		t.Fatalf("unexpected new coin event: %v, expected: %v", newCoinEvents[0], expected)
	}
}

func TestMintUtxoValueMismatchWithSats(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max":  uint64(21000000),
			"sats": uint64(10000),
		},
	}, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{
					Address: "1234",
					Value:   5000,
				},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405",
				},
			},
		},
	)

	if err == nil || err.Error() != "the valid output of Carv Coin CARV should be an integer multiple of 10000, tx = &{ [] [{5000 1234 } {0  OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405}]}" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMintTotalSupplyExceed(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 21000000,
		Args: map[string]interface{}{
			"max":  uint64(21000000),
			"sats": uint64(10000),
		},
	}, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vout: []mempool.Vout{
				{
					Address: "1234",
					Value:   10000,
				},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405",
				},
			},
		},
	)

	if err == nil || err.Error() != "mint Carv Coin CARV exceed max supply, totalSupply = 21000000, delta = 1, max = 21000000" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMintSuccess(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinsInUtxos([]string{}).Return(nil, nil)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max":  uint64(21000000),
			"sats": uint64(10000),
		},
	}, nil)

	newCoinEvents, balanceChangeEvents, err := carv.Parse(
		&mempool.Transaction{
			Txid: "5678",
			Vout: []mempool.Vout{
				{
					Address: "1234",
					Value:   10000,
				},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405",
				},
			},
		},
	)

	if err != nil || len(newCoinEvents) != 0 || len(balanceChangeEvents) != 1 {
		t.Fatal("should have found one balance change event")
	}

	expected := &types.BalanceChangeEvent{
		ChainId:  "bitcoin",
		Protocol: "carv",
		CoinId:   "CARV",
		Address:  "1234",
		Delta:    1,
		Utxo:     "5678:0",
		IsMint:   true,
	}
	if !reflect.DeepEqual(balanceChangeEvents[0], expected) {
		t.Fatalf("unexpected balance change event: %v, expected: %v", balanceChangeEvents[0], expected)
	}
}

func TestTransferUtxoValueNotMultiplyOfSatsError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max":  uint64(21000000),
			"sats": uint64(10000),
		},
	}, nil)
	mockDb.EXPECT().GetCoinsInUtxos([]string{"5678:0"}).Return([]*types.UnspentCoin{
		{
			CoinId: "CARV",
			Owner:  "1234",
			Amount: 1,
			Utxo:   "5678:0",
		},
	}, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vin: []mempool.Vin{
				{
					Txid: "5678",
					Vout: 0,
				},
			},
			Vout: []mempool.Vout{
				{
					Address: "1234",
					Value:   15000,
				},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405",
				},
			},
		},
	)

	if err == nil || err.Error() != "the valid output of Carv Coin CARV should be an integer multiple of 10000, tx = &{ [{5678 0}] [{15000 1234 } {0  OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405}]}" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransferInsufficientInputsError(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max":  uint64(21000000),
			"sats": uint64(10000),
		},
	}, nil)
	mockDb.EXPECT().GetCoinsInUtxos([]string{"5678:0"}).Return([]*types.UnspentCoin{
		{
			CoinId: "CARV",
			Owner:  "1234",
			Amount: 1,
			Utxo:   "5678:0",
		},
	}, nil)

	_, _, err := carv.Parse(
		&mempool.Transaction{
			Vin: []mempool.Vin{
				{
					Txid: "5678",
					Vout: 0,
				},
			},
			Vout: []mempool.Vout{
				{
					Address: "1234",
					Value:   10000,
				},
				{
					Address: "1234",
					Value:   10000,
				},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405",
				},
			},
		},
	)

	if err == nil || err.Error() != "insufficient inputs for transfer, input = 1, output = 2" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransferSuccess(t *testing.T) {
	mockDb, carv := setup(t)
	mockDb.EXPECT().GetCoinInfoById("CARV").Return(&types.CoinInfo{
		Id:          "CARV",
		TotalSupply: 1,
		Args: map[string]interface{}{
			"max":  uint64(21000000),
			"sats": uint64(10000),
		},
	}, nil)
	mockDb.EXPECT().GetCoinsInUtxos([]string{"5678:0"}).Return([]*types.UnspentCoin{
		{
			CoinId: "CARV",
			Owner:  "1234",
			Amount: 1,
			Utxo:   "5678:0",
		},
	}, nil)

	newCoinEvents, balanceChangeEvents, err := carv.Parse(
		&mempool.Transaction{
			Txid: "9abc",
			Vin: []mempool.Vin{
				{
					Txid: "5678",
					Vout: 0,
				},
			},
			Vout: []mempool.Vout{
				{
					Address: "1234",
					Value:   10000,
				},
				{
					Asm: "OP_RETURN OP_PUSHBYTES_1 43 OP_PUSHBYTES_3 82a405",
				},
			},
		},
	)

	if err != nil || len(newCoinEvents) != 0 || len(balanceChangeEvents) != 2 {
		t.Fatal("should have found two balance change events")
	}

	expected := []*types.BalanceChangeEvent{
		{
			ChainId:  "bitcoin",
			Protocol: "carv",
			CoinId:   "CARV",
			Address:  "1234",
			Delta:    -1,
			Utxo:     "5678:0",
		},
		{
			ChainId:  "bitcoin",
			Protocol: "carv",
			CoinId:   "CARV",
			Address:  "1234",
			Delta:    1,
			Utxo:     "9abc:0",
		},
	}
	if !reflect.DeepEqual(balanceChangeEvents, expected) {
		t.Fatalf("unexpected balance change events: %v, expected: %v", balanceChangeEvents, expected)
	}
}
