# Testing Matrix

Default tests are offline and require no Hyperion API access, Aptos fullnode
access, or API keys.

## Commands

```bash
make verify
```

`make verify` runs the same default checks expected before opening a PR:

- `make fmt-check`
- `make test`
- `make vet`

The raw Go commands remain useful when a narrower check is needed:

```bash
go test ./...
go vet ./...
```

If you use `mise`, the repository includes a local tool and task definition for
the same Go version and common verification commands:

```bash
mise trust
mise run verify
mise run test
mise run govulncheck
```

`mise` is optional local developer tooling. Hosted CI still uses
`actions/setup-go` with the Go version declared in `go.mod`.

Nix users can enter the pinned development shell from the repository flake:

```bash
nix develop
make verify
go test -count=1 ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

The flake pins `nixpkgs-unstable` to a revision that provides Go `1.25.11`,
matching `go.mod`. Nix is also optional local developer tooling; Hosted CI does
not depend on it.

## Hosted CI

GitHub Actions uses the Go toolchain declared in `go.mod` and runs the same
offline verification path for pull requests, pushes to `main`, and manual
`workflow_dispatch` runs:

- `make verify`
- `go test -count=1 ./...`
- `govulncheck ./...`

Hosted CI intentionally does not set `HYPERION_INTEGRATION=1`, so live Hyperion
REST and Aptos fullnode smoke tests stay disabled by default. This keeps PR
verification deterministic and independent of secrets, API keys, fullnode
availability, or upstream network conditions.

## Current Coverage

| Area | Coverage | Test files |
| --- | --- | --- |
| SDK initialization | Mainnet/testnet defaults, API key propagation, service handles, required option validation | `client_test.go` |
| Request layer | URL normalization, query encoding, JSON decoding, non-2xx errors | `request_test.go` |
| Utilities | Tick complement, slippage bounds, slippage calculation, currency checks, round tick, price-to-tick, tick-to-price | `utils_test.go` |
| Pool REST reads | Pools list, pool by ID, pool by token pair and fee tier, ticks, and typed stable-field decoding | `modules_test.go`, `rest_types_test.go` |
| Position REST reads | Positions by owner, ownership by position ID, non-zero fee history filtering, and typed stable-field decoding | `modules_test.go`, `rest_types_test.go` |
| Reward REST reads | Non-zero reward history filtering and typed stable-field decoding | `modules_test.go`, `rest_types_test.go` |
| Swap quotes | `flag=out`, `flag=in`, safe mode query behavior, and typed quote decoding | `modules_test.go`, `rest_types_test.go` |
| Aggregate route fetch | Mainnet-only guard and route response decoding | `aggregate_test.go` |
| Aggregate composer recorder and submit adapter boundary | Exact-in, exact-out refund, DEX adapter dispatch, mainnet guard, read-only plan handoff, submit adapter plan handoff, unsupported Go SDK gap, and error boundaries | `aggregate_composer_test.go`, `aggregate_submit_adapter_test.go` |
| Payload builders | Pool, legacy liquidity, router_v3 liquidity, pool estimate view payloads, position amount view payload, swap coin/FA argument order, reward, and claim argument order | `payload_test.go`, `router_v3_payload_test.go`, `pool_payload_test.go`, `position_payload_test.go` |
| Aptos view execution | REST `/v1/view` request conversion, API key header, status errors, client wiring, typed decoding errors, and service wrappers | `view_executor_test.go`, `view_services_test.go`, `typed_view_helpers_test.go` |
| Live integration harness | Environment parsing, default skips, client option construction, Aptos view smoke, and guarded Hyperion REST smoke tests | `integration_helpers_test.go`, `integration_support_test.go`, `view_integration_test.go`, `integration_hyperion_test.go` |
| Parity fixtures | Golden snapshots for representative payloads, including swap coin-type conversion, and REST responses | `parity_test.go`, `testdata/parity` |

## Exported API Coverage

| API surface | Coverage status | Evidence |
| --- | --- | --- |
| `Init`, `New`, `Options`, `InitOptions`, `Network` | Offline unit tests for defaults, required options, URL normalization, API key propagation, and service initialization | `client_test.go` |
| `Client.View`, `ViewExecutor`, `AptosViewExecutor`, `ViewStatusError` | Offline tests for missing executor errors, REST `/v1/view` conversion, API key header, versioned URL handling, and non-2xx errors; opt-in live timestamp smoke | `view_executor_test.go`, `view_integration_test.go` |
| `RequestClient`, `QueryParams`, `HTTPStatusError` | Offline tests for query encoding, JSON decoding, nil output handling, and non-2xx response details | `request_test.go` |
| Pool REST reads: `FetchAllPools`, `FetchPoolByID`, `GetPoolByTokenPairAndFeeTier`, `FetchTicks`, and `*Typed` variants | Offline unit tests, typed decode tests, and parity REST fixtures; opt-in live pool-list smoke | `modules_test.go`, `rest_types_test.go`, `parity_test.go`, `integration_hyperion_test.go` |
| Pool payload/view helpers: `CreatePoolTransactionPayload`, `CreateLiquiditySinglePayload`, `EstCurrencyAAmountFromBPayload`, `EstCurrencyBAmountFromAPayload`, typed optimal liquidity, pool state, and pool quote helpers | Offline payload/golden tests for builders and fake-executor tests for view wrappers | `payload_test.go`, `router_v3_payload_test.go`, `pool_payload_test.go`, `view_services_test.go`, `typed_view_helpers_test.go` |
| Position REST reads: `FetchAllPositionsByAddress`, `FetchPositionByID`, `FetchFeeHistory`, and `*Typed` variants | Offline unit tests for request paths, typed decode paths, and zero-amount fee filtering | `modules_test.go`, `rest_types_test.go` |
| Position payload/view helpers: `AddLiquidityTransactionPayload`, `AddLiquiditySinglePayload`, `AddLiquiditySingleCoinsPayload`, `RemoveLiquidityTransactionPayload`, `RemoveLiquidityMultiAgentDirectlyDepositPayload`, typed position value helpers, claim payload helpers | Offline payload/golden tests, strict recipient/multi-agent validation, and fake-executor view wrapper tests | `payload_test.go`, `router_v3_payload_test.go`, `position_payload_test.go`, `parity_test.go`, `view_services_test.go`, `typed_view_helpers_test.go` |
| Reward REST and payload/view helpers: `FetchRewardHistory`, `FetchRewardHistoryTyped`, `FetchRewardsPayload`, `FetchRewards`, `ClaimRewardPayload` | Offline unit tests for reward filtering, typed decode paths, payload construction, and fake-executor view wrapper behavior | `modules_test.go`, `rest_types_test.go`, `payload_test.go`, `view_services_test.go` |
| Swap quote methods: `EstFromAmount`, `EstToAmount`, `EstFromAmountTyped`, `EstToAmountTyped`, `EstimateAmountArgs` | Offline unit tests, typed decode tests, and parity REST fixtures; opt-in live quote smoke when token env is supplied | `modules_test.go`, `rest_types_test.go`, `parity_test.go`, `integration_hyperion_test.go` |
| Swap payload builders: `SwapTransactionPayload`, `SwapWithPartnershipTransactionPayload` | Offline tests for FA-to-FA, coin-to-coin conversion, original pair selection, partnership behavior, and golden parity snapshots | `payload_test.go`, `parity_test.go` |
| Aggregate route fetch: `EstAmountByAggregateSwap`, route/result structs, aggregate constants | Offline mainnet guard and parity route fixture; opt-in live aggregate route smoke on mainnet when token env is supplied | `aggregate_test.go`, `parity_test.go`, `integration_hyperion_test.go` |
| Aggregate composer: `GenerateAggregateSwapTransactionScript`, `BuildAggregateSwapSubmitPlan`, `BuildAggregateSwapSubmitTransaction`, `AggregateSwapComposer`, `AggregateSwapRecorder`, `AggregateSwapSubmitAdapter`, call argument/result types | Offline recorder tests for exact-in, exact-out refund, DEX adapter dispatch, read-only plan handoff, submit adapter handoff, unsupported adapter errors, composer errors, and mainnet guard | `aggregate_composer_test.go`, `aggregate_submit_adapter_test.go` |
| Utility functions and fee-tier constants | Offline tests for tick complement, pool deadline, slippage, currency validation, `LogBase`, price/tick conversion, tick rounding, fee-tier tables, and `U64Max` parity | `utils_test.go` |

## Golden Fixtures

Golden fixtures under `testdata/parity` are the migration calibration layer.
They intentionally snapshot the public payload and REST shapes that need to stay
aligned with the upstream TypeScript SDK.

Dynamic payload deadlines are normalized to `"<deadline>"` before comparison so
golden tests can focus on function names, type arguments, and stable argument
ordering.

## Future Test Work

| Area | Follow-up |
| --- | --- |
| Broader REST response structs | Expand typed wrappers when Hyperion publishes stable schemas for additional fields. |
| Native submit-ready aggregate adapter | Keep behind adapter-specific tests if upstream Aptos Go support gains a Dynamic Script Composer-compatible compiler. |

## Integration Tests

Live tests are opt-in. They are skipped unless `HYPERION_INTEGRATION=1` is set,
and individual scenarios skip again unless their required environment variables
are present.

| Variable | Required for | Notes |
| --- | --- | --- |
| `HYPERION_INTEGRATION=1` | All live tests | Without this, every `Integration` test skips. |
| `HYPERION_NETWORK` | All live tests | Optional `mainnet` or `testnet`; defaults to `mainnet` when omitted. |
| `HYPERION_API_HOST` | Hyperion REST override | Optional; defaults to the SDK's network API host. |
| `APTOS_FULLNODE_URL` | Aptos view smoke | Optional unless running `TestIntegrationAptosViewExecutorTimestamp`. |
| `APTOS_API_KEY` | Aptos view smoke | Optional; passed as a Bearer token when present. |
| `HYPERION_SWAP_FROM` | Swap quote smoke | Required with `HYPERION_SWAP_TO`; `HYPERION_SWAP_AMOUNT` defaults to `1000`. |
| `HYPERION_SWAP_TO` | Swap quote smoke | Required with `HYPERION_SWAP_FROM`. |
| `HYPERION_SWAP_AMOUNT` | Swap quote smoke | Optional amount override. |
| `HYPERION_AGG_FROM` | Aggregate route smoke | Mainnet only; required with `HYPERION_AGG_INPUT` and `HYPERION_AGG_TO`. |
| `HYPERION_AGG_INPUT` | Aggregate route smoke | Mainnet only. |
| `HYPERION_AGG_TO` | Aggregate route smoke | Mainnet only. |
| `HYPERION_AGG_AMOUNT` | Aggregate route smoke | Optional amount override, defaults to `1000`. |

Integration tests are not required by `go test ./...` in a clean local checkout
or CI environment.

Run all enabled live smoke tests with:

```bash
HYPERION_INTEGRATION=1 make test-integration
```

Run only the live Aptos view smoke test with:

```bash
HYPERION_INTEGRATION=1 \
APTOS_FULLNODE_URL=https://<aptos-fullnode>/v1 \
APTOS_API_KEY=<optional-key> \
go test -count=1 -run TestIntegrationAptosViewExecutorTimestamp ./...
```
