# Examples

These examples are numbered as a suggested learning path from SDK setup to
advanced aggregate swap composition.

| Step | Directory | Focus |
| --- | --- | --- |
| 001 | [001_init](001_init/main.go) | Initialize the SDK with mainnet/testnet defaults |
| 002 | [002_payloads](002_payloads/main.go) | Build deterministic offline transaction payloads |
| 003 | [003_read_pools](003_read_pools/main.go) | Read pool data through Hyperion REST endpoints |
| 004 | [004_swap_quote](004_swap_quote/main.go) | Request swap quotes |
| 005 | [005_view](005_view/main.go) | Execute Aptos view calls through `ViewExecutor` |
| 006 | [006_aggregate_composer](006_aggregate_composer/main.go) | Record aggregate swap composer calls |

Run an example with:

```bash
go run ./examples/001_init
```
