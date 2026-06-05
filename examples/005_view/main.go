package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

// This example executes a live Aptos `/v1/view` call through the SDK's
// AptosViewExecutor. It is different from the Hyperion REST quote example:
// `APTOS_FULLNODE_URL` must point to an Aptos fullnode, while the currency
// values are Hyperion/Aptos fungible-asset metadata addresses.
//
// The default pool-estimate parameters below are example mainnet values for a
// USD1 -> USDC pool from Hyperion REST pool metadata. Pool metadata can change;
// replace these values with the current pool you want to inspect for production
// usage.
//
// Copy this into your shell to run the example:
//
//	export APTOS_FULLNODE_URL=https://api.mainnet.aptoslabs.com/v1
//	# export APTOS_API_KEY=<optional-key>
//	export HYPERION_VIEW_CURRENCY_A=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2 # USD1
//	export HYPERION_VIEW_CURRENCY_B=0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b   # USDC
//	go run ./examples/005_view
//
// APTOS_API_KEY is optional for public fullnodes that allow anonymous access.
// When set, the SDK sends it as `Authorization: Bearer <key>` on the Aptos
// fullnode request. It does not sign transactions or change the view result; it
// is mainly for provider auth, higher/stable rate limits, usage tracking,
// billing, and revocation when a node provider requires or recommends a key.
const exampleEnv = `set APTOS_FULLNODE_URL, HYPERION_VIEW_CURRENCY_A, and HYPERION_VIEW_CURRENCY_B to execute a live Aptos view call

Example:
  export APTOS_FULLNODE_URL=https://api.mainnet.aptoslabs.com/v1
  # export APTOS_API_KEY=<optional-key>
  export HYPERION_VIEW_CURRENCY_A=0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2 # USD1
  export HYPERION_VIEW_CURRENCY_B=0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b   # USDC
  go run ./examples/005_view

Optional pool-estimate overrides for another pool:
  export HYPERION_VIEW_FEE_TIER_INDEX=0
  export HYPERION_VIEW_TICK_LOWER=-60
  export HYPERION_VIEW_TICK_UPPER=60
  export HYPERION_VIEW_CURRENT_PRICE_TICK=-13
  export HYPERION_VIEW_CURRENCY_B_AMOUNT=1000

HYPERION_VIEW_CURRENCY_B_AMOUNT is a base-unit amount. For 6-decimal tokens,
1000 means 0.001 token.

APTOS_API_KEY is optional for public fullnodes that allow anonymous access. If
set, the SDK sends it as Authorization: Bearer <key>. It does not sign
transactions or change the view result; it is mainly for provider auth,
higher/stable rate limits, usage tracking, billing, and key revocation.
`

const (
	defaultFeeTierIndex     = "0"
	defaultTickLower        = "-60"
	defaultTickUpper        = "60"
	defaultCurrentPriceTick = "-13"
	defaultCurrencyBAmount  = "1000"
)

func main() {
	fullnodeURL := os.Getenv("APTOS_FULLNODE_URL")
	if fullnodeURL == "" {
		fmt.Print(exampleEnv)
		return
	}
	currencyA := os.Getenv("HYPERION_VIEW_CURRENCY_A")
	currencyB := os.Getenv("HYPERION_VIEW_CURRENCY_B")
	if currencyA == "" || currencyB == "" {
		fmt.Print(exampleEnv)
		return
	}

	sdk, err := hyperion.Init(hyperion.InitOptions{
		Network:          hyperion.NetworkMainnet,
		AptosFullNodeURL: fullnodeURL,
		AptosAPIKey:      os.Getenv("APTOS_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	values, err := sdk.Pool.EstCurrencyAAmountFromB(context.Background(), hyperion.EstCurrencyAAmountFromBArgs{
		CurrencyA:        currencyA,
		CurrencyB:        currencyB,
		FeeTierIndex:     getenvDefault("HYPERION_VIEW_FEE_TIER_INDEX", defaultFeeTierIndex),
		TickLower:        getenvDefault("HYPERION_VIEW_TICK_LOWER", defaultTickLower),
		TickUpper:        getenvDefault("HYPERION_VIEW_TICK_UPPER", defaultTickUpper),
		CurrentPriceTick: getenvDefault("HYPERION_VIEW_CURRENT_PRICE_TICK", defaultCurrentPriceTick),
		CurrencyBAmount:  getenvDefault("HYPERION_VIEW_CURRENCY_B_AMOUNT", defaultCurrencyBAmount),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("view returned %d values: %v\n", len(values), values)
}

func getenvDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
