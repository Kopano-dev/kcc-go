# Kopano Core client for Go

This implements a minimal client interfacing with a couple of
SOAP methods of a Kopano server.

## Environment variables

| Environment variable       | Description                                   |
|----------------------------|-----------------------------------------------|
| KOPANO_SERVER_DEFAULT_URI  | URI used to connect to Kopano server          |
| TEST_USERNAME              | Kopano username used in unit tests            |
| TEST_PASSWORD              | Kopano username's password used in unit tests |

## Testing

Running the unit tests requires a Kopano Server with accessible SOAP service.
Make sure to set the environment variables as listed above to match your Kopano
server details.

```
go test -v
```

## Benchmark

For testing there is also a benchmark test.

```
go test -v -bench=. -run BenchmarkLogon -benchmem
BenchmarkLogon-8            2000            591907 ns/op           20509 B/op       217 allocs/op
PASS
ok      stash.kopano.io/kc/kcc-go       1.255s
```
