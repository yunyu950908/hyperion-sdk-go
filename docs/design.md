# Hyperion SDK Go Migration Design

This document tracks the migration from the upstream TypeScript SDK at
`Hyperionxyz/hyperion-sdk/packages/sdk` to the Go module
`github.com/yunyu950908/hyperion-sdk-go`.

## Upstream Snapshot

- Package: `@hyperionxyz/sdk`
- Version audited: `0.1.2`
- Source root: `packages/sdk/src`
- Public root: `src/index.ts`

The upstream `0.1.x` SDK moved Hyperion data reads from GraphQL to REST. The Go
SDK should not reintroduce removed low-level GraphQL APIs unless a future
requirement explicitly asks for them.

## Module Mapping

| TypeScript source | Go target | Notes |
| --- | --- | --- |
| `src/index.ts` | `Client`, `Options`, `Init` | Go uses explicit errors and service fields instead of JS getters. |
| `src/config/index.ts` | network constants and `InitOptions` | Mainnet/testnet contract addresses and API hosts are copied from upstream. |
| `src/modules/requestModule.ts` | `RequestClient` | Go adds `context.Context`, injectable `http.Client`, and typed HTTP status errors. |
| `src/modules/poolModule.ts` | `PoolService` | REST reads, create-pool payloads, pool estimate view payloads, and pool math helpers are ported. Live Aptos view execution remains separate. |
| `src/modules/positionModule.ts` | `PositionService` | REST reads, zero-amount filtering, liquidity payloads, claim payloads, strict remove-recipient validation, and amount-by-liquidity view payload construction are ported. Live view execution remains separate. |
| `src/modules/rewardModule.ts` | `RewardService` | Reward history, pending reward view payload, and claim payloads are ported. |
| `src/modules/swapModule.ts` | `SwapService` | Quote methods, basic payloads, swap coin-type to FA metadata conversion, aggregate route fetch, and aggregate composer recorder strategy are ported. |
| `src/utils/index.ts` | shared utility functions/constants | Tick complement, fee tier config, slippage helpers, price/tick conversion, and pool deadline helpers are implemented. |
| `src/helper/aggregateSwap/*` | aggregate route types/helper | Route fetch and composer call planning are ported through a Go composer interface and deterministic recorder. Submit-ready transaction serialization remains an adapter concern. |

## Go API Principles

- Every networked method accepts `context.Context`.
- HTTP behavior is testable through an injected `*http.Client`.
- REST endpoints decode into flexible `JSONMap` values until Hyperion publishes
  stable response schemas for every endpoint.
- Transaction payload builders should return typed Go structs instead of raw
  `map[string]any`.
- Numeric token amounts should avoid JavaScript-style precision loss. Where the
  upstream SDK returns JS numbers, Go payload builders may keep string values if
  that is safer for Aptos SDK consumers; differences must be documented.
- Aggregate swap route fetching is part of the core SDK. Aggregate transaction
  composition uses an `AggregateSwapComposer` interface plus
  `AggregateSwapRecorder` to mirror TypeScript batched-call planning without
  requiring a specific Aptos Go transaction builder.

## REST Endpoint Inventory

| Module | Method | Endpoint |
| --- | --- | --- |
| Pool | `FetchAllPools` | `GET /base/data/pools/stats` |
| Pool | `FetchPoolByID` | `GET /base/data/pools/stats?poolId=...` |
| Pool | `GetPoolByTokenPairAndFeeTier` | `GET /base/data/pools/by-token-pair` |
| Pool | `FetchTicks` | `GET /base/data/pools/{poolId}/liquidity-accumulation` |
| Position | `FetchAllPositionsByAddress` | `GET /base/data/positions?address=...` |
| Position | `FetchPositionByID` | `GET /base/data/liquidity/ownerships` |
| Position | `FetchFeeHistory` | `GET /base/data/rewards/claimed-fees` |
| Reward | `FetchRewardHistory` | `GET /base/data/rewards/claimed-farms` |
| Swap | `EstFromAmount` | `GET /base/rate/getSwapInfo?flag=out` |
| Swap | `EstToAmount` | `GET /base/rate/getSwapInfo?flag=in` |
| Swap | `EstAmountByAggregateSwap` | `GET /base/aggregator/getAggRoute` |

## Open Design Items

- Add live Aptos view execution for pool and position view payloads, tracked by
  #13. The current SDK builds parity view payloads but does not submit them to a
  fullnode.
- Add a submit-ready Aptos Go SDK adapter for `AggregateSwapComposer` if a
  production integration requires direct transaction serialization. The current
  recorder is intentionally deterministic and offline.
- Basic swap payloads emulate `aptos-tool` token-pair selection and
  `Token.faTypeCalculate()`: `address::module::name` values are treated as coin
  types, APT maps to `0xa`, and other coin types derive FA metadata with Aptos
  `createObjectAddress(0xa, shortCoinTypeBytes)`. Partnership swap payloads keep
  raw coin type arguments because the upstream TypeScript SDK does not apply the
  FA conversion in that builder.
- Replace `JSONMap` REST returns with stronger response structs if upstream API
  schemas become available.
