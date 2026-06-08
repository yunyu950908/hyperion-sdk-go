package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

// This example fetches a live aggregate route and converts it into an
// AggregateSwapSubmitPlan for wallet-layer handoff.
//
// The plan is intentionally not a wallet payload. A wallet or transaction layer
// still needs an external transaction composer, such as the upstream TypeScript
// Dynamic Script Composer path, to turn the recorded calls and result references
// into real Aptos transaction bytes before signing or submission.
const exampleEnv = `set HYPERION_AGG_FROM, HYPERION_AGG_INPUT, and HYPERION_AGG_TO to build an aggregate submit plan handoff

Example mainnet assetType values:
  USDC native, verified: 0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b
  USD1 native, verified: 0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2

Example:
  export HYPERION_AGG_FROM=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2  # USD1
  export HYPERION_AGG_INPUT=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2 # USD1
  export HYPERION_AGG_TO=0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b    # USDC
  export HYPERION_AGG_AMOUNT=1000
  export HYPERION_AGG_SLIPPAGE=0.5
  export HYPERION_AGG_PARTNERSHIP_ID=hyperion-aggregator
  go run ./examples/010_aggregate_plan_handoff

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
	partnershipID := getenvDefault("HYPERION_AGG_PARTNERSHIP_ID", hyperion.AggregatorPartnerName)

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

	plan, err := sdk.Swap.BuildAggregateSwapSubmitPlan(hyperion.BuildAggregateSwapSubmitPlanArgs{
		Route:         *route,
		PartnershipID: partnershipID,
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
		plan.RouteSplits,
		plan.RefundRouteSplits,
		firstDex(route),
	)
	fmt.Printf("submit plan calls=%d partnershipId=%q\n", len(plan.Calls), plan.PartnershipID)
	if len(plan.Calls) > 0 {
		fmt.Printf("first call: %s\n", plan.Calls[0].Function)
		fmt.Printf("last call: %s\n", plan.Calls[len(plan.Calls)-1].Function)
	}
	fmt.Println("handoff JSON follows; it is a composer plan, not submit-ready transaction bytes:")
	printJSON(plan)
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

func printJSON(value any) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encoded))
}
