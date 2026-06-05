package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hyperion "github.com/yunyu950908/hyperion-sdk-go"
)

func main() {
	sdk, err := hyperion.Init(hyperion.InitOptions{
		Network: hyperion.NetworkMainnet,
	})
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("HYPERION_EXAMPLE_LIVE") != "1" {
		fmt.Println("set HYPERION_EXAMPLE_LIVE=1 to fetch live pools")
		return
	}

	pools, err := sdk.Pool.FetchAllPools(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("loaded %d pools\n", len(pools))
}
