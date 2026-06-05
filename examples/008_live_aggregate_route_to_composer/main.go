package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

// This example connects the live aggregate route API to the aggregate composer.
//
// It fetches a real mainnet aggregate route, then records the batched call plan
// with AggregateSwapRecorder. The recorder output is still offline: it is not a
// submit-ready transaction, does not sign anything, and does not submit to Aptos.
const exampleEnv = `set HYPERION_AGG_FROM, HYPERION_AGG_INPUT, and HYPERION_AGG_TO to record a live aggregate route call plan

Example mainnet assetType values:
  USDC native, verified: 0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b
  USD1 native, verified: 0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2

Example:
  export HYPERION_AGG_FROM=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2  # USD1
  export HYPERION_AGG_INPUT=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2 # USD1
  export HYPERION_AGG_TO=0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b    # USDC
  export HYPERION_AGG_AMOUNT=1000
  export HYPERION_AGG_SLIPPAGE=0.5
  go run ./examples/008_live_aggregate_route_to_composer

Amounts are base-unit strings. For 6-decimal tokens, 1000 means 0.001 token.
`

func main() {
	from := os.Getenv("HYPERION_AGG_FROM")
	input := os.Getenv("HYPERION_AGG_INPUT")
	to := os.Getenv("HYPERION_AGG_TO")
	if from == "" || input == "" || to == "" {
		fmt.Print(exampleEnv)
		return
	}

	amount := getenvDefault("HYPERION_AGG_AMOUNT", "1000")
	slippage := getenvDefault("HYPERION_AGG_SLIPPAGE", "0.5")

	sdk, err := hyperion.Init(hyperion.InitOptions{
		Network: hyperion.NetworkMainnet,
	})
	if err != nil {
		log.Fatal(err)
	}

	route, err := sdk.Swap.EstAmountByAggregateSwap(context.Background(), hyperion.AggregateSwapRouteArgs{
		Amount:   amount,
		From:     from,
		Input:    input,
		Slippage: slippage,
		To:       to,
	})
	if err != nil {
		log.Fatal(err)
	}

	recorder := hyperion.NewAggregateSwapRecorder()
	err = sdk.Swap.GenerateAggregateSwapTransactionScript(hyperion.GenerateAggregateSwapTransactionScriptArgs{
		Route:    *route,
		Composer: recorder,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("aggregate exactIn=%t amountIn=%s minOut=%s quotedOut=%s\n",
		route.ExactIn,
		route.FromTokenAmount,
		route.MinToTokenAmount,
		route.ToTokenAmount,
	)
	fmt.Printf("route splits=%d refund splits=%d first DEX=%s\n",
		len(route.Quotes.Route),
		len(route.Quotes.RefundRoute),
		firstDex(route),
	)
	fmt.Printf("recorded %d aggregate composer calls\n", len(recorder.Calls))
	if len(recorder.Calls) > 0 {
		fmt.Printf("first call: %s\n", recorder.Calls[0].Function)
		fmt.Printf("last call: %s\n", recorder.Calls[len(recorder.Calls)-1].Function)
	}
	fmt.Println("this is an offline aggregate call plan only; it is not submitted transaction bytes")
}

func getenvDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func firstDex(route *hyperion.AggregateSwapInfoResult) string {
	for _, split := range route.Quotes.Route {
		for _, hop := range split.RouteTaken {
			if strings.TrimSpace(hop.DexName) != "" {
				return hop.DexName
			}
		}
	}
	return "<none>"
}
