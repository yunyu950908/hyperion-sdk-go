# Wallet Integration Guide

This SDK builds Hyperion reads, view helpers, payloads, and aggregate call plans.
It does not manage private keys, wallet authorization, transaction simulation, or
chain submission. Wallet-facing applications should treat SDK output by shape.

## Normal Entry-Function Payloads

Pool, position, reward, and normal swap payload builders return
`EntryFunctionPayload`. These payloads are unsigned transaction payloads. Pass
them to an Aptos wallet or transaction layer that can construct, simulate, sign,
and submit entry-function transactions.

Examples:

- `SwapTransactionPayload`
- `CreatePoolTransactionPayload`
- `AddLiquidityTransactionPayload`
- `RemoveLiquidityTransactionPayload`
- `ClaimFeeTransactionPayload`
- `ClaimRewardTransactionPayload`

`RemoveLiquidityMultiAgentDirectlyDepositPayload` returns
`MultiAgentPayloadEnvelope`. Its `Payload` is the entry-function payload, while
`SecondarySignerAddresses` must be passed to the downstream multi-agent
transaction builder. The signer addresses are intentionally not included in
`FunctionArguments`.

## Aggregate Route Handoff

Aggregate swaps are different from normal swap payloads. Hyperion returns an
aggregate route, and the SDK can convert that route into deterministic composer
calls:

```go
plan, err := sdk.Swap.BuildAggregateSwapSubmitPlan(hyperion.BuildAggregateSwapSubmitPlanArgs{
	Route:         *route,
	PartnershipID: hyperion.AggregatorPartnerName,
})
```

`AggregateSwapSubmitPlan` is a handoff object for a transaction composer. It is
not a wallet payload, not Move script bytecode, not BCS transaction bytes, and
not a signing message. A wallet cannot sign it directly.

Use the plan when your application has a separate composer layer, such as:

- the upstream TypeScript Dynamic Script Composer flow
- a TS/WASM bridge owned by the application
- a future native Go composer that can produce Aptos script payload bytes

If your application does not have that composer yet, keep aggregate swaps in a
quote/preview state or disable aggregate submit. Do not treat
`AggregateSwapRecorder` calls or `AggregateSwapSubmitPlan` JSON as a
submit-ready transaction.

## Adapter Boundary

When a composer is available, implement `AggregateSwapSubmitAdapter` and call:

```go
tx, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(ctx, hyperion.BuildAggregateSwapSubmitTransactionArgs{
	Route:         *route,
	PartnershipID: hyperion.AggregatorPartnerName,
	Adapter:       adapter,
})
```

The adapter is responsible for converting the deterministic plan into
adapter-specific transaction output, such as script payload bytes, BCS bytes, or
a signing message. The SDK still does not sign, simulate, or submit.

Use `NewUnsupportedAggregateSwapSubmitAdapter` in tests or feature gates to make
the current Go SDK gap explicit:

```go
_, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(ctx, hyperion.BuildAggregateSwapSubmitTransactionArgs{
	Route:   *route,
	Adapter: hyperion.NewUnsupportedAggregateSwapSubmitAdapter(),
})
```

## Preflight Checklist

Before showing a transaction to a wallet user, the application should verify:

- asset addresses and route endpoints match the user's intent
- amounts are raw base-unit integer strings, not floats
- slippage and any deadline policy are explicit
- recipient or signer semantics are visible to the user
- multi-agent secondary signers are passed through the wallet layer
- aggregate plans are converted by a trusted composer before signing
- simulation succeeds before submission when the wallet stack supports it

See `examples/007_swap_payload_from_quote` for a normal swap payload handoff and
`examples/010_aggregate_plan_handoff` for an aggregate plan handoff.
