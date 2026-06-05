package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

func main() {
	fullnodeURL := os.Getenv("APTOS_FULLNODE_URL")
	if fullnodeURL == "" {
		fmt.Println("set APTOS_FULLNODE_URL to execute a live Aptos view call")
		return
	}
	currencyA := os.Getenv("HYPERION_VIEW_CURRENCY_A")
	currencyB := os.Getenv("HYPERION_VIEW_CURRENCY_B")
	if currencyA == "" || currencyB == "" {
		fmt.Println("set HYPERION_VIEW_CURRENCY_A and HYPERION_VIEW_CURRENCY_B for the pool estimate view")
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
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyBAmount:  "1000",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("view returned %d values\n", len(values))
}
