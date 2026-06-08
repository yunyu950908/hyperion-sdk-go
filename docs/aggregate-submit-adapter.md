# Aggregate Submit Adapter Boundary

Aggregate swap routes need a multi-call transaction shape: one call can produce
a fungible-asset value that later calls borrow, copy, merge, or deposit. The SDK
therefore keeps two separate layers:

- `AggregateSwapRecorder` records the deterministic batched-call plan.
- `AggregateSwapSubmitAdapter` is an integration boundary for turning that plan
  into adapter-specific submit-ready transaction data.

## Current Go SDK Boundary

Current Aptos Go SDK releases can build ordinary entry-function transactions and
script transactions when compiled script bytecode is already available. They do
not expose the TypeScript Dynamic Script Composer compiler that converts
batched call/result references into Move script bytecode.

For that reason, this SDK does not ship a native submit-ready aggregate adapter
yet. The built-in `NewUnsupportedAggregateSwapSubmitAdapter()` returns a
structured `ErrAggregateSubmitAdapterUnsupported` error instead of treating the
recorder output as a transaction payload.

## Adapter Flow

Use `BuildAggregateSwapSubmitTransaction` when an application has its own
adapter, such as a TS/WASM bridge or a future native Go composer:

```go
result, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(ctx, hyperion.BuildAggregateSwapSubmitTransactionArgs{
	Route:         route,
	PartnershipID: "partner-id",
	Adapter:       adapter,
})
```

The SDK first composes the route through `AggregateSwapRecorder`, then passes an
`AggregateSwapSubmitPlan` to the adapter. The plan includes:

- the original aggregate route
- route/refund split counts
- partnership ID
- deterministic composer calls

The adapter returns `AggregateSwapSubmitTransaction`, which can carry BCS bytes,
a signing message, or an adapter-specific transaction object through `Raw`.

## Unsupported Adapter

Use the unsupported adapter to make the boundary explicit in tests or production
feature gates:

```go
_, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(ctx, hyperion.BuildAggregateSwapSubmitTransactionArgs{
	Route:   route,
	Adapter: hyperion.NewUnsupportedAggregateSwapSubmitAdapter(),
})
if errors.Is(err, hyperion.ErrAggregateSubmitAdapterUnsupported) {
	// Fall back to recorder inspection or a TypeScript Dynamic Script Composer path.
}
```

## Non-Goals

- No private key handling.
- No wallet authorization.
- No simulation or chain submission.
- No script bytecode generation from recorder calls without compiler support.
