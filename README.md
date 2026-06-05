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
- swap quote requests
- transaction payload builders for common pool, liquidity, reward, and swap
  workflows
- swap coin-type to fungible-asset metadata conversion for normal swap payloads
- aggregate route fetching and a deterministic aggregate composer recorder
- Aptos view execution through an injectable `ViewExecutor`
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

Networked read methods accept `context.Context` and return flexible `JSONMap`
values for REST endpoints whose public schemas can still change.

```go
pools, err := sdk.Pool.FetchAllPools(ctx)
positions, err := sdk.Position.FetchAllPositionsByAddress(ctx, "0xowner")
rewards, err := sdk.Reward.FetchRewardHistory(ctx, "position-id", "0xowner")
```

See [examples/read_pools](examples/read_pools/main.go) for a compile-checked
read example.

## Swap Quotes

```go
quote, err := sdk.Swap.EstFromAmount(ctx, hyperion.EstimateAmountArgs{
	Amount:   "1000000",
	From:     "0x...",
	To:       "0x...",
	SafeMode: true,
})
```

See [examples/swap_quote](examples/swap_quote/main.go).

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

See [examples/payloads](examples/payloads/main.go).

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

`EntryFunctionPayload` preserves the upstream TypeScript SDK's camelCase JSON
field names for parity snapshots. The REST view executor converts it to Aptos
fullnode request fields: `function`, `type_arguments`, and `arguments`.

See [examples/view](examples/view/main.go). The example only executes when
`APTOS_FULLNODE_URL`, `HYPERION_VIEW_CURRENCY_A`, and
`HYPERION_VIEW_CURRENCY_B` are set.

## Aggregate Swap

`EstAmountByAggregateSwap` fetches aggregate routes on mainnet. The SDK also
provides an `AggregateSwapComposer` interface and a built-in
`AggregateSwapRecorder`.

The recorder mirrors the upstream TypeScript SDK's batched-call order for tests,
audits, and future adapter development. It does not serialize submit-ready Aptos
transaction bytes.

```go
recorder := hyperion.NewAggregateSwapRecorder()
err := sdk.Swap.GenerateAggregateSwapTransactionScript(hyperion.GenerateAggregateSwapTransactionScriptArgs{
	Route:    route,
	Composer: recorder,
})
```

See [examples/aggregate_composer](examples/aggregate_composer/main.go).

## Testing

Default tests do not require live network access or credentials.

```bash
make verify
go test ./...
go vet ./...
```

Opt-in live smoke tests are documented in [docs/testing.md](docs/testing.md).

## Known Boundaries

- REST read methods return `JSONMap` until Hyperion publishes stable response
  schemas suitable for strong Go structs.
- Aggregate composer support records deterministic batched-call plans; a
  submit-ready Aptos transaction adapter is a future extension.
- Removed low-level GraphQL APIs from older TypeScript SDK versions are not
  recreated as first-class Go APIs.

## More Documentation

- [Migration notes](docs/migration.md)
- [Design notes](docs/design.md)
- [Testing matrix](docs/testing.md)
