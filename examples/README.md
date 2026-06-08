# Examples

These examples are numbered as a suggested learning path from SDK setup to
payload construction, aggregate routes, and position operations.

| Step | Directory | Focus |
| --- | --- | --- |
| 001 | [001_init](001_init/main.go) | Initialize the SDK with mainnet/testnet defaults |
| 002 | [002_payloads](002_payloads/main.go) | Build deterministic offline transaction payloads |
| 003 | [003_read_pools](003_read_pools/main.go) | Read pool data through Hyperion REST endpoints |
| 004 | [004_swap_quote](004_swap_quote/main.go) | Request typed swap quotes |
| 005 | [005_view](005_view/main.go) | Execute Aptos view calls through `ViewExecutor` |
| 006 | [006_aggregate_composer](006_aggregate_composer/main.go) | Record aggregate swap composer calls |
| 007 | [007_swap_payload_from_quote](007_swap_payload_from_quote/main.go) | Build an unsigned normal swap payload from a live quote path |
| 008 | [008_live_aggregate_route_to_composer](008_live_aggregate_route_to_composer/main.go) | Convert a live aggregate route into an offline composer call plan |
| 009 | [009_position_liquidity_payloads](009_position_liquidity_payloads/main.go) | Build position and liquidity payloads without submitting transactions |

Run an example with:

```bash
go run ./examples/001_init
```
