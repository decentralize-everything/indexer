package extract

type Vin interface {
	GetTxid() string
	GetVout() int
}

type Vout interface {
	GetValue() float64
	GetAddress() string
	GetAsm() string
}

type Transaction interface {
	GetTxid() string
	GetVin() []Vin
	GetVout() []Vout
}

type Block interface {
	GetHash() string
	GetTime() int
	GetHeight() int
	GetTxs() []Transaction
}
