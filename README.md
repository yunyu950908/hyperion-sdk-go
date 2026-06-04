# hyperion-sdk-go

Go SDK for Hyperion on Aptos.

This repository ports the upstream TypeScript SDK at
`Hyperionxyz/hyperion-sdk/packages/sdk` to Go module path:

```bash
go get github.com/yunyu950908/hyperion-sdk-go
```

## Status

The migration is tracked in GitHub issues, with the roadmap issue at
[#10](https://github.com/yunyu950908/hyperion-sdk-go/issues/10).

Implemented so far:

- SDK initialization with mainnet/testnet defaults.
- REST request layer with `context.Context`, injectable `http.Client`, and typed
  HTTP status errors.
- Pool REST reads.
- Position and reward REST reads with upstream zero-amount filtering behavior.
- Swap quote REST methods.
- Aggregate swap route fetch with the upstream mainnet-only guard and a
  recorder-backed aggregate composer strategy for deterministic batched-call
  plans.
- Entry-function payload builders for common pool, position, reward, and swap
  workflows, including basic swap coin-type to fungible-asset metadata
  conversion.
- Pool math helpers for price/tick conversion and pool estimate view payload
  builders.
- Position and reward payload builders, including amount-by-liquidity view
  payload construction.

Still tracked:

- Aptos view-call integration.
- Live transaction integration for aggregate composer adapters.
- Strong response structs for REST endpoints if stable API schemas become
  available.

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
		Network:     hyperion.NetworkMainnet,
		AptosAPIKey: "",
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

## Custom Client

```go
sdk, err := hyperion.New(hyperion.Options{
	Network:                   hyperion.NetworkTestnet,
	ContractAddress:           hyperion.TestnetContractAddress,
	HyperionFullNodeIndexerURL: "https://api-testnet.hyperion.xyz",
	HyperionAPIHost:           "https://api-testnet.hyperion.xyz",
	AptosAPIKey:               "",
})
```

Legacy upstream URLs ending in `/v1/graphql` are normalized to the REST API host.

## Swap Quotes

```go
quote, err := sdk.Swap.EstFromAmount(context.Background(), hyperion.EstimateAmountArgs{
	Amount:   "1000000",
	From:     "0x...",
	To:       "0x...",
	SafeMode: true,
})
```

## Aggregate Swap Composer

Aggregate swap route fetching is available through `EstAmountByAggregateSwap`.
`GenerateAggregateSwapTransactionScript` accepts an `AggregateSwapComposer`
interface; the built-in recorder mirrors the upstream TypeScript SDK's
batched-call order without serializing a submit-ready Aptos transaction.

```go
route, err := sdk.Swap.EstAmountByAggregateSwap(ctx, hyperion.AggregateSwapRouteArgs{
	Amount:   "1000000",
	From:     "0x...",
	Input:    "0x...",
	Slippage: "0.5",
	To:       "0x...",
})
if err != nil {
	return err
}

recorder := hyperion.NewAggregateSwapRecorder()
err = sdk.Swap.GenerateAggregateSwapTransactionScript(hyperion.GenerateAggregateSwapTransactionScriptArgs{
	Route:    *route,
	Composer: recorder,
})
```

## Payload Builders

Payload builders return `EntryFunctionPayload`, matching the upstream SDK's
entry-function shape while preserving amount fields as strings to avoid
JavaScript-style precision loss.

```go
payload, err := sdk.Swap.SwapTransactionPayload(hyperion.SwapTransactionPayloadArgs{
	CurrencyA:       "0x1::aptos_coin::AptosCoin", // coin type input is accepted
	CurrencyB:       "0x...",
	CurrencyAAmount: "1000000",
	CurrencyBAmount: "990000",
	Slippage:        "0.5",
	PoolRoute:       []string{"0xpool"},
	Recipient:       "0xrecipient",
})
```

## Testing

Default tests do not require live network access or credentials.

```bash
go test ./...
go vet ./...
```

See [docs/testing.md](docs/testing.md) for the coverage matrix, parity fixture
layout, and integration test plan.

## Design Notes

See [docs/design.md](docs/design.md) for the TypeScript-to-Go module mapping and
known migration gaps.
