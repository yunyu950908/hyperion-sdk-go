package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

// Mainnet assetType examples from Hyperion REST pool metadata:
//
//	USDT bridged, verified:
//	  0xe568e9322107a5c9ba4cbd05a630a5586aa73e744ada246c3efb0f4ce3e295f3
//	USDC native, verified:
//	  0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b
//	USD1 native, verified:
//	  0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2
//
// Copy one pair into your shell before running this example:
//
//	export HYPERION_SWAP_FROM=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2 # USD1
//	export HYPERION_SWAP_TO=0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b   # USDC
//	export HYPERION_SWAP_AMOUNT=1000
//	go run ./examples/004_swap_quote
const exampleEnv = `set HYPERION_SWAP_FROM and HYPERION_SWAP_TO to request a live quote

Example mainnet assetType values:
  USDT bridged, verified: 0xe568e9322107a5c9ba4cbd05a630a5586aa73e744ada246c3efb0f4ce3e295f3
  USDC native, verified:  0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b
  USD1 native, verified:  0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2

Example:
  export HYPERION_SWAP_FROM=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2 # USD1
  export HYPERION_SWAP_TO=0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b   # USDC
  export HYPERION_SWAP_AMOUNT=1000
  go run ./examples/004_swap_quote
`

func main() {
	from := os.Getenv("HYPERION_SWAP_FROM")
	to := os.Getenv("HYPERION_SWAP_TO")
	if from == "" || to == "" {
		fmt.Print(exampleEnv)
		return
	}

	amount := os.Getenv("HYPERION_SWAP_AMOUNT")
	if amount == "" {
		amount = "1000"
	}

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
	fmt.Printf("quote flag=%s amountOut=%s path pools=%d\n",
		quote.Flag,
		quote.ResolvedAmountOut(),
		len(quote.Path),
	)
}
