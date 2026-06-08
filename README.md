# hyperion-sdk-go

Go SDK for Hyperion on Aptos.

This repository ports the upstream TypeScript SDK at
`Hyperionxyz/hyperion-sdk/packages/sdk` to Go module path
`github.com/yunyu950908/hyperion-sdk-go`.

## Installation

```bash
go get github.com/yunyu950908/hyperion-sdk-go
```

## Status

The migration roadmap is tracked in
[#10](https://github.com/yunyu950908/hyperion-sdk-go/issues/10). The SDK now
covers the main TypeScript SDK surface:

- mainnet/testnet initialization
- Hyperion REST reads for Pool, Position, Reward, and Swap
- additive typed REST response wrappers for selected stable fields
- swap quote requests
- transaction payload builders for common pool, liquidity, reward, and swap
  workflows
- swap coin-type to fungible-asset metadata conversion for normal swap payloads
- aggregate route fetching and a deterministic aggregate composer recorder
- Aptos view execution through an injectable `ViewExecutor`
- typed Aptos view helpers for position, pool, quote, price hub, protocol guard,
  and coin wrapper reads
- parity fixtures, exported API coverage docs, `make verify`, and opt-in live
  integration smoke tests

## Quick Start

```go
package main

import (
	"context"
	"log"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

func main() {
	sdk, err := hyperion.Init(hyperion.InitOptions{
		Network: hyperion.NetworkMainnet,
	})
	if err != nil {
		log.Fatal(err)
	}

	pools, err := sdk.Pool.FetchAllPools(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %d pools", len(pools))
}
```

## Examples

The examples are numbered as a suggested learning path:

| Step | Example | Focus |
| --- | --- | --- |
| 001 | [examples/001_init](examples/001_init/main.go) | SDK initialization and network defaults |
| 002 | [examples/002_payloads](examples/002_payloads/main.go) | Offline transaction payload builders |
| 003 | [examples/003_read_pools](examples/003_read_pools/main.go) | Hyperion REST pool reads |
| 004 | [examples/004_swap_quote](examples/004_swap_quote/main.go) | Typed swap quote requests |
| 005 | [examples/005_view](examples/005_view/main.go) | Aptos view calls through `ViewExecutor` |
| 006 | [examples/006_aggregate_composer](examples/006_aggregate_composer/main.go) | Aggregate swap composer recording |
| 007 | [examples/007_swap_payload_from_quote](examples/007_swap_payload_from_quote/main.go) | Build an unsigned normal swap payload from a live quote path |
| 008 | [examples/008_live_aggregate_route_to_composer](examples/008_live_aggregate_route_to_composer/main.go) | Convert a live aggregate route into an offline composer call plan |
| 009 | [examples/009_position_liquidity_payloads](examples/009_position_liquidity_payloads/main.go) | Build position and liquidity payloads without submitting transactions |
| 010 | [examples/010_aggregate_plan_handoff](examples/010_aggregate_plan_handoff/main.go) | Build an aggregate submit plan handoff for wallet-layer composers |

## Initialization

Use `Init` for the built-in Hyperion mainnet/testnet defaults:

```go
sdk, err := hyperion.Init(hyperion.InitOptions{
	Network: hyperion.NetworkMainnet,
})
```

Use `New` when you need explicit hosts, a custom `http.Client`, or an Aptos
fullnode URL for live view calls:

```go
sdk, err := hyperion.New(hyperion.Options{
	Network:                   hyperion.NetworkTestnet,
	ContractAddress:           hyperion.TestnetContractAddress,
	HyperionFullNodeIndexerURL: "https://api-testnet.hyperion.xyz",
	HyperionAPIHost:           "https://api-testnet.hyperion.xyz",
	AptosFullNodeURL:          "https://<aptos-fullnode>/v1",
	AptosAPIKey:               "",
})
```

Legacy upstream URLs ending in `/v1/graphql` are normalized to the REST API host.

## Common Reads

Networked read methods accept `context.Context`. Legacy REST read methods keep
returning flexible `JSONMap` values for callers that need every dynamic
upstream field.

```go
pools, err := sdk.Pool.FetchAllPools(ctx)
positions, err := sdk.Position.FetchAllPositionsByAddress(ctx, "0xowner")
rewards, err := sdk.Reward.FetchRewardHistory(ctx, "position-id", "0xowner")
```

Typed wrappers are also available for selected stable fields. They preserve
integer amounts and ticks as strings and intentionally omit unstable REST fields.

```go
pools, err := sdk.Pool.FetchAllPoolsTyped(ctx)
positions, err := sdk.Position.FetchAllPositionsByAddressTyped(ctx, "0xowner")
fees, err := sdk.Position.FetchFeeHistoryTyped(ctx, "position-id", "0xowner")
```

See [examples/003_read_pools](examples/003_read_pools/main.go) for a compile-checked
read example.

## Swap Quotes

```go
quote, err := sdk.Swap.EstFromAmountTyped(ctx, hyperion.EstimateAmountArgs{
	Amount:   "1000000",
	From:     "0x...",
	To:       "0x...",
	SafeMode: true,
})
amountOut := quote.ResolvedAmountOut()
poolRoute := quote.Path
```

See [examples/004_swap_quote](examples/004_swap_quote/main.go).

## Payload Builders

Payload builders return `EntryFunctionPayload`, matching the upstream SDK's
entry-function shape while preserving amount fields as strings to avoid
JavaScript-style precision loss.

```go
payload, err := sdk.Swap.SwapTransactionPayload(hyperion.SwapTransactionPayloadArgs{
	CurrencyA:       "0x1::aptos_coin::AptosCoin",
	CurrencyB:       "0x...",
	CurrencyAAmount: "1000000",
	CurrencyBAmount: "990000",
	Slippage:        "0.5",
	PoolRoute:       []string{"0xpool"},
	Recipient:       "0xrecipient",
})
```

See [examples/002_payloads](examples/002_payloads/main.go).

The SDK also exposes current `router_v3` liquidity payload builders for wallet
or Aptos transaction-layer integration:

```go
payload, err := sdk.Pool.CreateLiquiditySinglePayload(hyperion.CreateLiquiditySinglePayloadArgs{
	CurrencyA:            "0x...",
	CurrencyB:            "0x...",
	FeeTierIndex:         "0",
	TickLower:            "-3",
	TickUpper:            "-1",
	Amount:               "25086537",
	SlippageNumerator:    "99",
	SlippageDenominator:  "100",
	ThresholdNumerator:   "1",
	ThresholdDenominator: "1",
})
```

`RemoveLiquidityMultiAgentDirectlyDepositPayload` returns a
`MultiAgentPayloadEnvelope`: signer arguments are not included in
`FunctionArguments`; downstream signing code must use `SecondarySignerAddresses`
when constructing the Aptos multi-agent transaction. These builders do not sign,
simulate, or submit transactions.

## Aptos View Calls

Payload builders stay offline and deterministic. To execute an Aptos view,
configure `AptosFullNodeURL` or inject your own `ViewExecutor`.

```go
sdk, err := hyperion.Init(hyperion.InitOptions{
	Network:          hyperion.NetworkMainnet,
	AptosFullNodeURL: "https://<aptos-fullnode>/v1",
	AptosAPIKey:      "<optional-key>",
})
if err != nil {
	return err
}

values, err := sdk.Pool.EstCurrencyAAmountFromB(ctx, hyperion.EstCurrencyAAmountFromBArgs{
	CurrencyA:        "0x...",
	CurrencyB:        "0x...",
	FeeTierIndex:     "2",
	TickLower:        "-60",
	TickUpper:        "60",
	CurrentPriceTick: "0",
	CurrencyBAmount:  "1000000",
})
```

Typed view wrappers are available for position values, optimal liquidity, pool
state, quote helpers, price hub reads, protocol guard checks, and coin wrapper
identity reads:

```go
amounts, err := sdk.Position.FetchPositionTokenAmounts(ctx, "0xposition")
poolInfo, err := sdk.Position.FetchPositionPoolInfo(ctx, "0xposition")
quote, err := sdk.Swap.FetchBatchAmountOut(ctx, hyperion.BatchAmountOutArgs{
	PoolRoute: []string{"0xpool"},
	AmountIn:  "1000000",
	TokenIn:   "0x...",
	TokenOut:  "0x...",
})
```

Price hub helpers expose on-chain price preview/source comparison without
converting `u256` or `u64` values to floats:

```go
preview, err := sdk.PriceHub.FetchPricePreview(ctx, hyperion.PricePreviewArgs{
	Asset:  "0x...",       // FA metadata object address
	Amount: "1000000000", // raw base-unit integer string
})
comparison, err := sdk.PriceHub.FetchPriceSourceComparison(ctx, "0x...")
```

Protocol guard helpers expose `rate_limiter_check` views so wallets and
services can preflight limiter state before simulation or submission:

```go
status, err := sdk.RateLimiter.FetchGlobalAssetRateLimiter(ctx, "0x...")
poolGuard, err := sdk.RateLimiter.FetchPoolUPriceLimiter(ctx, "0xpool")
globalGuard, err := sdk.RateLimiter.FetchGlobalUPriceLimiter(ctx)
```

Coin wrapper helpers normalize token identity across coin type strings and FA
metadata object addresses:

```go
isWrapper, err := sdk.CoinWrapper.FetchIsWrapper(ctx, "0x...")
coinType, err := sdk.CoinWrapper.FetchCoinType(ctx, "0x...")
formatted, err := sdk.CoinWrapper.FetchFormattedFungibleAsset(ctx, "0x...")
```

These helpers are read-only: they do not sign, simulate, submit, bypass protocol
guards, or decide wallet policy. Use their typed output as transaction preflight
or UI/service diagnostics, then pass any transaction payload to your wallet or
Aptos transaction layer.

`EntryFunctionPayload` preserves the upstream TypeScript SDK's camelCase JSON
field names for parity snapshots. The REST view executor converts it to Aptos
fullnode request fields: `function`, `type_arguments`, and `arguments`.

See [examples/005_view](examples/005_view/main.go). The example only executes when
`APTOS_FULLNODE_URL`, `HYPERION_VIEW_CURRENCY_A`, and
`HYPERION_VIEW_CURRENCY_B` are set.

## Aggregate Swap

`EstAmountByAggregateSwap` fetches aggregate routes on mainnet. The SDK also
provides an `AggregateSwapComposer` interface and a built-in
`AggregateSwapRecorder`.

The recorder mirrors the upstream TypeScript SDK's batched-call order for tests
and audits. It does not serialize submit-ready Aptos transaction bytes.
`BuildAggregateSwapSubmitPlan` returns the same deterministic
`AggregateSwapSubmitPlan` directly, so wallet-facing applications can hand the
plan to their own transaction composer without pretending the plan is a wallet
payload. For applications that provide their own TS/WASM bridge or future native
Go composer, `BuildAggregateSwapSubmitTransaction` delegates transaction
construction to an `AggregateSwapSubmitAdapter`.

```go
recorder := hyperion.NewAggregateSwapRecorder()
err := sdk.Swap.GenerateAggregateSwapTransactionScript(hyperion.GenerateAggregateSwapTransactionScriptArgs{
	Route:    route,
	Composer: recorder,
})
```

```go
plan, err := sdk.Swap.BuildAggregateSwapSubmitPlan(hyperion.BuildAggregateSwapSubmitPlanArgs{
	Route:         route,
	PartnershipID: "partner-id",
})
```

Current Aptos Go SDK releases can wrap compiled script bytecode, but they do not
expose the TypeScript Dynamic Script Composer compiler that turns batched
call/result references into Move script bytecode. Use
`NewUnsupportedAggregateSwapSubmitAdapter` to make that gap explicit, or inject a
real adapter when your application has one:

```go
_, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(ctx, hyperion.BuildAggregateSwapSubmitTransactionArgs{
	Route:   route,
	Adapter: hyperion.NewUnsupportedAggregateSwapSubmitAdapter(),
})
```

See [examples/006_aggregate_composer](examples/006_aggregate_composer/main.go).

See [examples/008_live_aggregate_route_to_composer](examples/008_live_aggregate_route_to_composer/main.go)
for the full route-fetch to recorder flow. See
[examples/010_aggregate_plan_handoff](examples/010_aggregate_plan_handoff/main.go)
for the route-fetch to submit-plan handoff flow. See
[docs/aggregate-submit-adapter.md](docs/aggregate-submit-adapter.md) for adapter
boundaries and current upstream limitations, and
[docs/wallet-integration.md](docs/wallet-integration.md) for wallet-layer
handoff guidance.

## Testing

Default tests do not require live network access or credentials.

```bash
make verify
go test ./...
go vet ./...
```

Opt-in live smoke tests are documented in [docs/testing.md](docs/testing.md).

## Known Boundaries

- Legacy REST read methods return `JSONMap`; typed REST wrappers expose selected
  stable fields and intentionally omit dynamic upstream fields.
- Typed Aptos view helpers decode selected on-chain view responses, but still
  require `AptosFullNodeURL` or an injected `ViewExecutor`.
- Price hub, rate limiter, and coin wrapper helpers are read-only protocol views;
  they do not submit transactions or override protocol guard behavior.
- Coin wrapper helpers accept fungible-asset metadata object addresses. Use
  coin-type strings such as `0x1::aptos_coin::AptosCoin` only where a method
  explicitly asks for a coin type.
- Hyperion contract ABI support and Hyperion UI support can differ; for example,
  `router_v3::add_liquidity_single` remains exposed by ABI even if a UI hides
  add zap-in flows.
- Aggregate composer support records deterministic batched-call plans, exposes
  them for wallet-layer handoff, and can pass them to an injected submit adapter;
  the built-in unsupported adapter documents the current Aptos Go SDK Dynamic
  Script Composer gap.
- Numeric public API fields intentionally stay as strings for chain-facing
  integer values; see [docs/numeric-helpers.md](docs/numeric-helpers.md) for the
  current `apd/v3` decision.
- Removed low-level GraphQL APIs from older TypeScript SDK versions are not
  recreated as first-class Go APIs.

## More Documentation

- [Migration notes](docs/migration.md)
- [Design notes](docs/design.md)
- [Wallet integration guide](docs/wallet-integration.md)
- [Testing matrix](docs/testing.md)
