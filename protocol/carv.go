package protocol

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/decentralize-everything/indexer/extract"
	"github.com/decentralize-everything/indexer/store"
	"github.com/decentralize-everything/indexer/types"
	"github.com/decentralize-everything/indexer/utils"
	"go.uber.org/zap"
)

var (
	CARV_PREFIX         = "OP_RETURN OP_PUSHBYTES_1 43 "
	COIN_ID_LEN_MIN     = 1
	COIN_ID_LEN_MAX     = 6
	COIN_SUPPLY_MIN     = uint64(1)
	COIN_SATS_MIN       = uint64(10000)
	COIN_MINT_LIMIT_MIN = uint64(1)
	COIN_LOCKED_BTC_MAX = uint64(21_000_000 * 1_000_000) // 1% of total BTC supply.
)

type CarvProtocol struct {
	db     store.Database
	logger *zap.Logger
}

var _ Parser = (*CarvProtocol)(nil)

func NewCarvProtocol(db store.Database, logger *zap.Logger) *CarvProtocol {
	return &CarvProtocol{
		db:     db,
		logger: logger,
	}
}

func (p *CarvProtocol) Parse(tx extract.Transaction) ([]*types.NewCoinEvent, []*types.BalanceChangeEvent, error) {
	var newCoinEvents []*types.NewCoinEvent
	var balanceChangeEvents []*types.BalanceChangeEvent

	// Find out all the burnt coins.
	utxos := make([]string, 0, len(tx.GetVin()))
	for _, vin := range tx.GetVin() {
		utxos = append(utxos, vin.GetTxid()+":"+strconv.Itoa(vin.GetVout()))
	}
	coins, err := p.db.GetCoinsInUtxos(utxos)
	if err != nil {
		return nil, nil, err
	}
	for _, coin := range coins {
		balanceChangeEvents = append(balanceChangeEvents, &types.BalanceChangeEvent{
			ChainId:  "bitcoin",
			Protocol: "carv",
			CoinId:   coin.CoinId,
			Address:  coin.Owner,
			Delta:    -coin.Amount,
			Utxo:     coin.Utxo,
		})
	}

	// Only one Carv protocol metadata is allowed per transaction.
	metaFound := false
	for i, vout := range tx.GetVout() {
		// Basic criteria for Carv protocol metadata.
		if vout.GetValue() != 0 || len(vout.GetAddress()) != 0 || !strings.HasPrefix(vout.GetAsm(), CARV_PREFIX) {
			continue
		}

		if metaFound {
			return nil, nil, fmt.Errorf("multiple Carv protocol metadata found in tx: %v", tx)
		}
		metaFound = true

		if len(vout.GetAsm()) < len(CARV_PREFIX+"OP_PUSHBYTES_") {
			return nil, nil, fmt.Errorf("metadata is too short: %s", vout.GetAsm())
		}

		segments := strings.Split(vout.GetAsm()[len(CARV_PREFIX+"OP_PUSHBYTES_"):], " ")
		if len(segments) != 2 {
			return nil, nil, fmt.Errorf("invalid metadata format: %s", vout.GetAsm())
		}

		length, err := strconv.Atoi(segments[0])
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing metadata length: %s", vout.GetAsm())
		}

		if len(segments[1]) != length*2 {
			return nil, nil, fmt.Errorf("metadata length mismatch: %s", vout.GetAsm())
		}

		meta, err := hex.DecodeString(segments[1])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode metadata into bytes: %s", vout.GetAsm())
		}
		args := utils.VarintDecodeArray(meta)

		if len(args) == 4 { // Deploy.
			id, max, sats, limit := utils.Base26Decode(args[0]), args[1], args[2], args[3]
			// Easy checks go first.
			if len(id) < COIN_ID_LEN_MIN || len(id) > COIN_ID_LEN_MAX || max < COIN_SUPPLY_MIN || sats < COIN_SATS_MIN || limit < COIN_MINT_LIMIT_MIN {
				return nil, nil, fmt.Errorf("invalid arguments for deployment, id = %s, max = %d, sats = %d, limit = %d", id, max, sats, limit)
			}

			lockedBtc := max * sats
			if lockedBtc/max != sats || lockedBtc > COIN_LOCKED_BTC_MAX { // Handle overflow.
				return nil, nil, fmt.Errorf("locked BTC out of range, max = %d, sats = %d, lockedBtc = %d", max, sats, lockedBtc)
			}

			if i != 0 {
				return nil, nil, fmt.Errorf("metadata for deployment placed at the %d-th UTXO, should be the first", i+1)
			}

			// Check if the coin ID is already taken.
			if ci, _ := p.db.GetCoinInfoById(id); ci != nil {
				return nil, nil, fmt.Errorf("coin ID already taken: %s", id)
			}

			newCoinEvents = append(newCoinEvents, &types.NewCoinEvent{
				ChainId:  "bitcoin",
				Protocol: "carv",
				CoinId:   id,
				Args: map[string]interface{}{
					"max":   max,
					"sats":  sats,
					"limit": limit,
				},
			})
		} else if len(args) == 1 { // Mint or transfer.
			id := utils.Base26Decode(args[0])
			ci, err := p.db.GetCoinInfoById(id)
			if err != nil || ci == nil {
				return nil, nil, fmt.Errorf("coin ID not found: %s", id)
			}

			totalInput := 0
			for _, coin := range coins {
				if coin.CoinId == id {
					totalInput += coin.Amount
				}
			}

			if totalInput == 0 { // Mint.
				// There must be exactly one valid UTXO following the metadata of the Carv protocol.
				if i != 1 || len(tx.GetVout()[0].GetAddress()) == 0 {
					return nil, nil, fmt.Errorf("invalid UTXO following mint metadata: %v", tx)
				}

				if uint64(tx.GetVout()[0].GetValue())%ci.Args["sats"].(uint64) != 0 {
					return nil, nil, fmt.Errorf("the valid output of Carv Coin %s should be an integer multiple of %d, tx = %v", id, ci.Args["sats"].(uint64), tx)
				}

				delta := uint64(tx.GetVout()[0].GetValue()) / ci.Args["sats"].(uint64)
				if uint64(ci.TotalSupply)+delta > ci.Args["max"].(uint64) {
					return nil, nil, fmt.Errorf("mint Carv Coin %s exceed max supply, totalSupply = %d, delta = %d, max = %d", id, ci.TotalSupply, delta, ci.Args["max"].(uint64))
				}

				balanceChangeEvents = append(balanceChangeEvents, &types.BalanceChangeEvent{
					ChainId:  "bitcoin",
					Protocol: "carv",
					CoinId:   id,
					Address:  tx.GetVout()[0].GetAddress(),
					Delta:    int(delta),
					Utxo:     tx.GetTxid() + ":0",
					IsMint:   true,
				})
			} else { // Transfer.
				totalOutput := uint64(0)
				for j := 0; j < i; j++ {
					vout := tx.GetVout()[j]
					if vout.GetValue() == 0 || uint64(vout.GetValue())%ci.Args["sats"].(uint64) != 0 || len(vout.GetAddress()) == 0 {
						return nil, nil, fmt.Errorf("the valid output of Carv Coin %s should be an integer multiple of %d, tx = %v", id, ci.Args["sats"].(uint64), tx)
					}
					totalOutput += uint64(vout.GetValue()) / ci.Args["sats"].(uint64)
					balanceChangeEvents = append(balanceChangeEvents, &types.BalanceChangeEvent{
						ChainId:  "bitcoin",
						Protocol: "carv",
						CoinId:   id,
						Address:  vout.GetAddress(),
						Delta:    int(uint64(vout.GetValue()) / ci.Args["sats"].(uint64)),
						Utxo:     tx.GetTxid() + ":" + strconv.Itoa(j),
					})
				}

				if totalInput < int(totalOutput) {
					return nil, nil, fmt.Errorf("insufficient inputs for transfer, input = %d, output = %d", totalInput, totalOutput)
				}
			}
		} else {
			return nil, nil, fmt.Errorf("invalid Carv protocol metadata: %s", vout.GetAsm())
		}
	}
	return newCoinEvents, balanceChangeEvents, nil
}
