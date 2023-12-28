# HTTP API v1

## Get balances of address

```shell
GET /api/v1/addresses/:address

eg. localhost:8080/api/v1/addresses/addr1

{
	"data": {
		"TESTCA": 1,
		"TESTCB": 3
	},
	"result": true
}
```

## Get coins of address

```shell
GET /api/v1/addresses/:address/coins

eg. localhost:8080/api/v1/addresses/addr1/coins

{
	"data": [
		{
			"CoinId": "TESTCA",
			"Owner": "addr1",
			"Amount": 1,
			"Utxo": "1111:0"
		},
		{
			"CoinId": "TESTCB",
			"Owner": "addr1",
			"Amount": 3,
			"Utxo": "1113:0"
		}
	],
	"result": true
}
```

## Get coin list

```shell
GET /api/v1/coins

params:
    * page >= 1, default 1
    * page_size <= 100, default 10
    * sorted_by tx_count(default)/holder_count/created_at
    * dir desc(default)/asc 

eg. localhost:8080/api/v1/coins

{
	"data": [
		{
			"Id": "TESTCL",
			"TotalSupply": 5,
			"Args": {
				"max": 100,
				"sats": 10000,
                "limit": 1
			},
			"TxCount": 12,
			"HolderCount": 89,
			"CreatedAt": 800004
		},
        ...
	],
	"result": true
}
```

## Get coin info

```shell
GET /api/v1/coins/:id

eg. localhost:8080/api/v1/coins/TESTCA

{
	"data": {
		"Id": "TESTCA",
		"TotalSupply": 5,
		"Args": {
			"max": 100,
			"sats": 10000,
            "limit": 1
		},
		"TxCount": 1,
		"HolderCount": 100,
		"CreatedAt": 800005
	},
	"result": true
}
```

# Run unit tests

```shell
go test ./... -coverprofile=coverage.out
```

# Check test coverage

```shell
go tool cover -html=coverage.out
```