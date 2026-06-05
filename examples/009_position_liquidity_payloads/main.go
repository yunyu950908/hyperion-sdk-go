package main

import (
	"encoding/json"
	"fmt"
	"log"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

// This example demonstrates offline position and liquidity payload builders.
//
// It does not read live ownership, sign with a wallet, or submit anything to
// Aptos. Real add/remove liquidity operations need an existing position ID,
// token amounts in base units, a strict recipient account address for removal,
// and an external wallet or Aptos transaction layer to sign and submit the
// generated payload.
func main() {
	sdk, err := hyperion.Init(hyperion.InitOptions{
		Network: hyperion.NetworkMainnet,
	})
	if err != nil {
		log.Fatal(err)
	}

	const (
		usd1            = "0x05fabd1b12e39967a3c24e91b7b8f67719a6dacee74f3c8b9fb7d93e855437d2"
		usdc            = "0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b"
		positionID      = "0xexample-position-id"
		strictRecipient = "0x0000000000000000000000000000000000000000000000000000000000000001"
	)

	createPool, err := sdk.Pool.CreatePoolTransactionPayload(hyperion.CreatePoolTransactionPayloadArgs{
		CurrencyA:        usd1,
		CurrencyB:        usdc,
		CurrencyAAmount:  "1000",
		CurrencyBAmount:  "1000",
		FeeTierIndex:     "0",
		CurrentPriceTick: "-13",
		TickLower:        "-60",
		TickUpper:        "60",
		Slippage:         "0.5",
	})
	if err != nil {
		log.Fatal(err)
	}

	addLiquidity, err := sdk.Position.AddLiquidityTransactionPayload(hyperion.AddLiquidityTransactionPayloadArgs{
		PositionID:      positionID,
		CurrencyA:       usd1,
		CurrencyB:       usdc,
		CurrencyAAmount: "1000",
		CurrencyBAmount: "1000",
		Slippage:        "0.5",
		FeeTierIndex:    "0",
	})
	if err != nil {
		log.Fatal(err)
	}

	removeLiquidity, err := sdk.Position.RemoveLiquidityTransactionPayload(hyperion.RemoveLiquidityTransactionPayloadArgs{
		PositionID:      positionID,
		CurrencyA:       usd1,
		CurrencyB:       usdc,
		CurrencyAAmount: "1000",
		CurrencyBAmount: "1000",
		DeltaLiquidity:  "12345",
		Slippage:        "0.5",
		Recipient:       strictRecipient,
	})
	if err != nil {
		log.Fatal(err)
	}

	payloads := map[string]hyperion.EntryFunctionPayload{
		"createPool":      createPool,
		"addLiquidity":    addLiquidity,
		"removeLiquidity": removeLiquidity,
		"claimFee":        sdk.Position.ClaimFeeTransactionPayload(positionID, strictRecipient),
		"claimReward":     sdk.Position.ClaimRewardTransactionPayload(positionID, strictRecipient),
		"claimAllRewards": sdk.Position.ClaimAllRewardsTransactionPayload(positionID, strictRecipient),
	}

	fmt.Println("offline position/liquidity payloads, ready for wallet/Aptos transaction signing:")
	printJSON(payloads)
}

func printJSON(value any) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encoded))
}
