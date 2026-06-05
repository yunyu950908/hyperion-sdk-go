package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

func main() {
	from := os.Getenv("HYPERION_SWAP_FROM")
	to := os.Getenv("HYPERION_SWAP_TO")
	if from == "" || to == "" {
		fmt.Println("set HYPERION_SWAP_FROM and HYPERION_SWAP_TO to request a live quote")
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

	quote, err := sdk.Swap.EstFromAmount(context.Background(), hyperion.EstimateAmountArgs{
		Amount:   amount,
		From:     from,
		To:       to,
		SafeMode: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("quote fields: %d\n", len(quote))
}
