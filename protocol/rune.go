package protocol

import (
	"fmt"
	"strings"

	"github.com/decentralize-everything/indexer/extract"
	"github.com/decentralize-everything/indexer/types"
	"go.uber.org/zap"
)

var (
	RUNE_PREFIX = "OP_RETURN OP_PUSHBYTES_1 52 "
)

type RuneProtocol struct {
	logger *zap.Logger
}

var _ Parser = (*RuneProtocol)(nil)

func NewRuneProtocol(logger *zap.Logger) *RuneProtocol {
	return &RuneProtocol{
		logger: logger,
	}
}

func (p *RuneProtocol) Parse(tx extract.Transaction) ([]*types.NewCoinEvent, []*types.BalanceChangeEvent, error) {
	for _, vout := range tx.GetVout() {
		if strings.HasPrefix(vout.GetAsm(), RUNE_PREFIX) {
			p.logger.Debug("found RUNE_PREFIX", zap.String("txid", tx.GetTxid()), zap.String("vout", fmt.Sprintf("%v", vout)))
		}
	}
	return nil, nil, nil
}
