# HTTP API v1

## Get indexer status

```shell
GET /api/v1/status

eg. localhost:8080/api/v1/status

{
	"data": {
		"indexed_height": 2568302,
		"network": "testnet"
	},
	"result": true
}
```

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
	"data": {
		"list": [
			{
				"Id": "PSBTS",
				"TotalSupply": 1,
				"Args": {
					"limit": 1000,
					"max": 21000000,
					"sats": 10000
				},
				"TxCount": 2,
				"HolderCount": 0,
				"CreatedAt": 1703823964,
				"DeployTx": "1234567890",
				"DeployHeight": 2567909
			},
            ...
		],
		"total": 13
	},
	"result": true
}
```

## Get coin info

```shell
GET /api/v1/coins/:id

eg. localhost:8080/api/v1/coins/TESTCA

{
    "data": {
        "Id": "PSBTS",
        "TotalSupply": 1,
        "Args": {
            "limit": 1000,
            "max": 21000000,
            "sats": 10000
        },
        "TxCount": 2,
        "HolderCount": 0,
        "CreatedAt": 1703823964,
        "DeployTx": "1234567890",
        "DeployHeight": 2567909
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