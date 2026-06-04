# Testing Matrix

Default tests are offline and require no Hyperion API access, Aptos fullnode
access, or API keys.

## Commands

```bash
go test ./...
go vet ./...
```

## Current Coverage

| Area | Coverage | Test files |
| --- | --- | --- |
| SDK initialization | Mainnet/testnet defaults, API key propagation, service handles, required option validation | `client_test.go` |
| Request layer | URL normalization, query encoding, JSON decoding, non-2xx errors | `request_test.go` |
| Utilities | Tick complement, slippage bounds, slippage calculation, currency checks, round tick, price-to-tick, tick-to-price | `utils_test.go` |
| Pool REST reads | Pools list, pool by ID, pool by token pair and fee tier, ticks | `modules_test.go` |
| Position REST reads | Positions by owner, ownership by position ID, non-zero fee history filtering | `modules_test.go` |
| Reward REST reads | Non-zero reward history filtering | `modules_test.go` |
| Swap quotes | `flag=out`, `flag=in`, safe mode query behavior | `modules_test.go` |
| Aggregate route fetch | Mainnet-only guard and route response decoding | `aggregate_test.go` |
| Payload builders | Pool, liquidity, pool estimate view payloads, position amount view payload, swap, reward, and claim argument order | `payload_test.go`, `pool_payload_test.go`, `position_payload_test.go` |
| Parity fixtures | Golden snapshots for representative payloads and REST responses | `parity_test.go`, `testdata/parity` |

## Golden Fixtures

Golden fixtures under `testdata/parity` are the migration calibration layer.
They intentionally snapshot the public payload and REST shapes that need to stay
aligned with the upstream TypeScript SDK.

Dynamic payload deadlines are normalized to `"<deadline>"` before comparison so
golden tests can focus on function names, type arguments, and stable argument
ordering.

## Known Gaps

| Gap | Tracking issue |
| --- | --- |
| Aptos view-call integration | #13 |
| Swap coin type to fungible asset metadata conversion | #6 |
| Aggregate swap transaction script composition | #7 |
| Strong REST response structs if schemas stabilize | #8 |
| Live integration tests behind environment variables | #8 |

## Integration Test Plan

Live tests should stay opt-in. When added, they should be skipped unless all
required environment variables are present.

Proposed environment variables:

- `HYPERION_INTEGRATION=1`
- `HYPERION_NETWORK=mainnet|testnet`
- `HYPERION_API_HOST`
- `APTOS_API_KEY`

Integration tests should never be required by `go test ./...` in a clean local
checkout or CI environment.
