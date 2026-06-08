package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

// This example connects the live quote API to the offline swap payload builder.
//
// It fetches a Hyperion REST quote, reads the returned pool path, and builds an
// EntryFunctionPayload with SwapTransactionPayload. The payload is not signed or
// submitted here; pass it to a wallet or an Aptos transaction layer to execute a
// real swap.
const exampleEnv = `set HYPERION_SWAP_FROM, HYPERION_SWAP_TO, and HYPERION_SWAP_RECIPIENT to build a swap payload from a live quote

Example mainnet assetType values:
  USDT bridged, verified: 0xe568e9322107a5c9ba4cbd05a630a5586aa73e744ada246c3efb0f4ce3e295f3
  USDC native, verified:  0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b
  USD1 native, verified:  0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2

Example:
  export HYPERION_SWAP_FROM=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2 # USD1
  export HYPERION_SWAP_TO=0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b   # USDC
  export HYPERION_SWAP_AMOUNT=1000
  export HYPERION_SWAP_SLIPPAGE=0.5
  export HYPERION_SWAP_RECIPIENT=0x0000000000000000000000000000000000000000000000000000000000000001
  go run ./examples/007_swap_payload_from_quote

Amounts are base-unit strings. For 6-decimal tokens, 1000 means 0.001 token.
`

func main() {
	from := os.Getenv("HYPERION_SWAP_FROM")
	to := os.Getenv("HYPERION_SWAP_TO")
	recipient := os.Getenv("HYPERION_SWAP_RECIPIENT")
	if from == "" || to == "" || recipient == "" {
		fmt.Print(exampleEnv)
		return
	}

	amount := getenvDefault("HYPERION_SWAP_AMOUNT", "1000")
	slippage := getenvDefault("HYPERION_SWAP_SLIPPAGE", "0.5")

	sdk, err := hyperion.Init(hyperion.InitOptions{
		Network: hyperion.NetworkMainnet,
	})
	if err != nil {
		log.Fatal(err)
	}

	quote, err := sdk.Swap.EstFromAmountTyped(context.Background(), hyperion.EstimateAmountArgs{
		Amount:   amount,
		From:     from,
		To:       to,
		SafeMode: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	path := quote.Path
	if len(path) == 0 {
		log.Fatal("quote did not include a pool path; use a direct Hyperion route quote before building a normal swap payload")
	}
	amountOut := quote.ResolvedAmountOut()
	if amountOut == "" {
		log.Fatal("quote did not include amountOut/outputAmount")
	}

	payload, err := sdk.Swap.SwapTransactionPayload(hyperion.SwapTransactionPayloadArgs{
		CurrencyA:       from,
		CurrencyB:       to,
		CurrencyAAmount: amount,
		CurrencyBAmount: amountOut,
		Slippage:        slippage,
		PoolRoute:       path,
		Recipient:       recipient,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("quote amountIn=%s amountOut=%s fee=%s path pools=%d\n",
		quote.ResolvedAmountIn(),
		amountOut,
		firstNonEmpty(quote.Fee, quote.FeeAmount),
		len(path),
	)
	fmt.Println("swap payload, ready for wallet/Aptos transaction signing:")
	printJSON(payload)
}

func getenvDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func printJSON(value any) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encoded))
}
