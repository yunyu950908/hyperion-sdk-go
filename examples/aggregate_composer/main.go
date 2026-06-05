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

	route := hyperion.AggregateSwapInfoResult{
		ExactIn:          true,
		FromTokenAmount:  "1000",
		MinToTokenAmount: "990",
		FromToken:        hyperion.TokenAddressInfo{Address: "0x1"},
		ToToken:          hyperion.TokenAddressInfo{Address: "0x2"},
		Quotes: hyperion.Quotes{
			Route: []hyperion.AggregateSwapRoute{
				{
					AmountIn:  "1000",
					AmountOut: "995",
					RouteTaken: []hyperion.AggregateSwapRouteTaken{
						{
							DexName:        "Hyperion",
							FromToken:      hyperion.TokenTypeInfo{TokenType: "0x1"},
							ToToken:        hyperion.TokenTypeInfo{TokenType: "0x2"},
							PoolID:         "0x1",
							A2B:            true,
							SqrtPriceLimit: "0",
							AmountIn:       "1000",
							AmountOut:      "995",
						},
					},
				},
			},
		},
	}

	recorder := hyperion.NewAggregateSwapRecorder()
	err = sdk.Swap.GenerateAggregateSwapTransactionScript(hyperion.GenerateAggregateSwapTransactionScriptArgs{
		Route:    route,
		Composer: recorder,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("recorded %d aggregate composer calls\n", len(recorder.Calls))
}
