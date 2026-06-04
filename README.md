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
- Aggregate swap route fetch with the upstream mainnet-only guard.
- Entry-function payload builders for common pool, position, reward, and swap
  workflows.

Still tracked:

- Aptos view-call integration.
- Aggregate swap transaction script composition.
- Swap payload coin-type to fungible-asset conversion. Current swap payload
  builders return an explicit error for coin types instead of silently producing
  a wrong payload.
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

## Payload Builders

Payload builders return `EntryFunctionPayload`, matching the upstream SDK's
entry-function shape while preserving amount fields as strings to avoid
JavaScript-style precision loss.

```go
payload, err := sdk.Swap.SwapTransactionPayload(hyperion.SwapTransactionPayloadArgs{
	CurrencyA:       "0x...",
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

## Design Notes

See [docs/design.md](docs/design.md) for the TypeScript-to-Go module mapping and
known migration gaps.
