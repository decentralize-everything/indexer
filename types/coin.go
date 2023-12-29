package types

import (
	"bytes"
	"encoding/gob"
)

type CoinInfo struct {
	Id           string
	TotalSupply  int
	Args         map[string]interface{}
	TxCount      int
	HolderCount  int
	CreatedAt    int
	DeployTx     string
	DeployHeight int
}

func (m *CoinInfo) ToBytes() []byte {
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(m); err != nil {
		panic(err)
	}
	return data.Bytes()
}

func (m *CoinInfo) FromBytes(bs []byte) error {
	return gob.NewDecoder(bytes.NewReader(bs)).Decode(m)
}

type UnspentCoin struct {
	CoinId string
	Owner  string
	Amount int
	Utxo   string
}

func (m *UnspentCoin) ToBytes() []byte {
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(m); err != nil {
		panic(err)
	}
	return data.Bytes()
}

func (m *UnspentCoin) FromBytes(bs []byte) error {
	return gob.NewDecoder(bytes.NewReader(bs)).Decode(m)
}
