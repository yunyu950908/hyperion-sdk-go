// Package hyperion provides a Go SDK for Hyperion on Aptos.
//
// The package is a Go port of the upstream TypeScript SDK. It exposes a root
// Client with Pool, Position, Reward, and Swap services, plus shared helpers for
// request handling, Hyperion constants, utility math, transaction payload
// construction, and Aptos view execution.
//
// Networked read methods accept context.Context and return explicit errors.
// Original REST read methods use JSONMap while Hyperion response schemas remain
// flexible; additive typed wrappers decode selected stable fields. Payload
// builders return EntryFunctionPayload values that mirror the upstream
// TypeScript SDK shape and can be used offline.
//
// Live Aptos view calls are available through Client.View, service convenience
// methods, and the ViewExecutor interface. The built-in AptosViewExecutor posts
// to an Aptos fullnode REST endpoint. Aggregate swap composition uses an
// AggregateSwapComposer interface and the deterministic AggregateSwapRecorder;
// the recorder does not serialize submit-ready transaction bytes.
//
// See README.md, docs/migration.md, docs/design.md, and docs/testing.md in the
// repository for usage examples, migration notes, design boundaries, and
// verification commands.
package hyperion
