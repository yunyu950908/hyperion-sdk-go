package main

import (
	"fmt"
	"log"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

func main() {
	sdk, err := hyperion.Init(hyperion.InitOptions{
		Network: hyperion.NetworkMainnet,
	})
	if err != nil {
		log.Fatal(err)
	}

	swap, err := sdk.Swap.SwapTransactionPayload(hyperion.SwapTransactionPayloadArgs{
		CurrencyA:       "0x1::aptos_coin::AptosCoin",
		CurrencyB:       "0x2",
		CurrencyAAmount: "1000",
		CurrencyBAmount: "990",
		Slippage:        "0.5",
		PoolRoute:       []string{"0x1"},
		Recipient:       "0x1",
	})
	if err != nil {
		log.Fatal(err)
	}

	fee := sdk.Position.ClaimFeeTransactionPayload("position-id", "0x1")
	reward := sdk.Reward.ClaimRewardPayload("position-id", "0x1")

	fmt.Println(swap.Function)
	fmt.Println(fee.Function)
	fmt.Println(reward.Function)
}
