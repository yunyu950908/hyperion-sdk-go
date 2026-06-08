# TypeScript SDK to Go Migration Notes

This document summarizes how the audited TypeScript SDK surface maps to
`github.com/yunyu950908/hyperion-sdk-go`, and which differences are intentional.

## API Mapping

| TypeScript SDK concept | Go SDK equivalent | Notes |
| --- | --- | --- |
| `initHyperionSDK(options)` | `hyperion.Init(hyperion.InitOptions)` | Uses built-in mainnet/testnet defaults and returns `(*Client, error)`. |
| `new HyperionSDK(options)` | `hyperion.New(hyperion.Options)` | Use for explicit hosts, custom HTTP clients, or view executor injection. |
| `HyperionSDK` | `Client` | The Go client exposes service fields instead of JavaScript getters. |
| `sdk.Pool`, `sdk.Position`, `sdk.Reward`, `sdk.Swap` | `Client.Pool`, `Client.Position`, `Client.Reward`, `Client.Swap` | Services hold methods for REST reads, payload builders, and view wrappers. |
| TypeScript REST reads | Go methods with `context.Context` | Networked methods accept a context and return explicit errors. |
| Dynamic REST objects | `JSONMap` plus typed wrappers | Existing REST methods keep flexible `JSONMap` returns; additive `*Typed` methods decode selected stable fields. |
| Entry-function payload builders | `EntryFunctionPayload` and typed args structs | Builders are deterministic and offline. |
| Aptos view calls | `Client.View` and `ViewExecutor` | Users can use the built-in REST executor or inject their own implementation. |
| Aggregate swap helper | `EstAmountByAggregateSwap` plus `AggregateSwapComposer` | Route fetching is supported; submit-ready transaction serialization is an adapter concern. |

## Intentional Differences

### Context and Errors

Go networked methods accept `context.Context` and return explicit `error`
values. This replaces JavaScript promise rejection and lets callers control
timeouts, cancellation, and request lifetimes.

### HTTP Injection

`Options.HTTPClient` lets tests and applications inject a custom `*http.Client`.
The request layer is covered with offline `httptest` fixtures.

### REST Response Shape

The upstream TypeScript SDK returns dynamic REST objects. The Go SDK mirrors
that with `JSONMap` on the original REST methods so callers can still access
every upstream field. Additive typed wrappers are available for selected stable
fields on pool, position, reward, and swap quote reads. These wrappers keep
integer amounts and ticks as strings and intentionally omit dynamic or unstable
REST fields.

### Token Amount Precision

Payload builders preserve token amounts as strings where that avoids
JavaScript-style precision loss. Callers should pass integer token amounts as
decimal strings.

### Coin Type Conversion

Normal swap payloads emulate the upstream `aptos-tool` token-pair selection and
coin type to fungible-asset metadata conversion:

- `0x1::aptos_coin::AptosCoin` maps to `0xa`
- other `address::module::name` coin types derive metadata through the Aptos
  object-address algorithm
- partnership swap payloads preserve raw coin type arguments, matching upstream
  behavior

### Aptos View Execution

`EntryFunctionPayload` keeps the upstream TypeScript SDK's `typeArguments` and
`functionArguments` JSON field names for parity snapshots. The built-in
`AptosViewExecutor` converts to the Aptos REST `/v1/view` request shape:
`type_arguments` and `arguments`.

### Aggregate Composer Boundary

The TypeScript SDK uses an Aptos script-composer dependency. The Go SDK exposes
an `AggregateSwapComposer` interface and a deterministic `AggregateSwapRecorder`
so callers can inspect or adapt the batched-call plan. The recorder does not
produce submit-ready transaction bytes.

### GraphQL

The audited upstream `0.1.x` SDK moved Hyperion data reads from GraphQL to REST.
Removed low-level GraphQL entry points are not recreated as first-class Go APIs.

## Verification

Default verification stays offline:

```bash
make verify
go test -count=1 ./...
```

Live integration smoke tests are opt-in and documented in
[docs/testing.md](testing.md).

## Future Extensions

- add a submit-ready aggregate composer adapter if a production Aptos Go
  transaction builder integration is required
- expand typed REST wrappers if Hyperion publishes broader stable schemas
