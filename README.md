# Run unit tests

```shell
go test ./... -coverprofile=coverage.out
```

# Check test coverage

```shell
go tool cover -html=coverage.out
```