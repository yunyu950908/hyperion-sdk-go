# Numeric Helper Strategy

This document records the current numeric policy for the SDK and the decision
from issue #41 about `github.com/cockroachdb/apd/v3`.

## Current Policy

The SDK keeps chain-facing amounts, liquidity, fees, price hub values, and rate
limit values as strings at public API boundaries. This avoids precision loss for
large Aptos integers and keeps wallet integrations close to the values returned
by REST and view APIs.

Current numeric work is intentionally narrow:

- `CheckSlippage`, `SlippageCalculator`, `roundDecimalString`, and zero-amount
  filtering use `math/big.Rat` for exact decimal parsing and integer rounding.
- price/tick helpers use `float64` because that path is calibrated against the
  upstream TypeScript SDK behavior.
- payload builders validate raw integer strings instead of converting them into
  a public decimal type.
- typed REST and typed view helpers preserve raw integer-like values as strings.

## apd/v3 Evaluation

`cockroachdb/apd/v3` is a strong arbitrary-precision decimal package. As of this
evaluation, the latest module version is `v3.2.3`.

Useful properties:

- `Decimal` represents arbitrary-precision decimal values.
- arithmetic is centralized through `Context`, which manages precision, range,
  rounding, condition flags, and traps.
- standard functions such as `sqrt`, `ln`, and `pow` are available.
- SQL scan/value helpers are available for database boundaries.

Those properties would be useful if the SDK grows a larger decimal calculation
surface. They are not currently necessary for the small number of exact decimal
operations in this repository.

## Decision

Do not introduce `cockroachdb/apd/v3` globally right now.

Keep the current design:

- public API numeric values remain strings
- slippage stays on the small `math/big.Rat` helper path
- price/tick behavior continues to prioritize upstream TypeScript parity
- no new decimal dependency is added until there is a broader calculation need

## Re-Evaluation Triggers

Revisit `apd/v3` when one or more of these conditions is true:

- multiple modules need user-input decimal parsing, display rounding, token
  decimal conversion, quote safety margins, or similar repeated logic
- rounding mode, precision, inexact, overflow, or trap policy needs to be
  centralized and audited
- SQL decimal storage or an external decimal data source becomes part of the SDK
- price/tick behavior intentionally moves from TypeScript parity toward
  high-precision decimal behavior

## Adoption Rules If Needed Later

If `apd/v3` is introduced later:

- keep `apd.Decimal` out of public SDK structs and method signatures
- add a small internal helper layer instead of scattering `Context` setup
- record precision, rounding mode, and trap policy in this document
- migrate one behavior at a time and keep existing parity tests green
