package hyperion

import "testing"

func TestInitUsesNetworkDefaultsAndAPIKey(t *testing.T) {
	t.Parallel()

	sdk, err := Init(InitOptions{
		Network:     NetworkMainnet,
		AptosAPIKey: "aptos-key",
	})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if sdk.Options.Network != NetworkMainnet {
		t.Fatalf("network = %q, want %q", sdk.Options.Network, NetworkMainnet)
	}
	if sdk.Options.ContractAddress != MainnetContractAddress {
		t.Fatalf("contract address = %q, want %q", sdk.Options.ContractAddress, MainnetContractAddress)
	}
	if sdk.Options.HyperionAPIHost != "https://api.hyperion.xyz" {
		t.Fatalf("api host = %q", sdk.Options.HyperionAPIHost)
	}
	if sdk.Options.AptosAPIKey != "aptos-key" {
		t.Fatalf("aptos api key = %q", sdk.Options.AptosAPIKey)
	}
	if sdk.Pool == nil || sdk.Position == nil || sdk.Reward == nil || sdk.Swap == nil {
		t.Fatalf("expected all SDK services to be initialized")
	}
}

func TestNewNormalizesLegacyGraphQLIndexerURL(t *testing.T) {
	t.Parallel()

	sdk, err := New(Options{
		Network:                    NetworkTestnet,
		ContractAddress:            TestnetContractAddress,
		HyperionFullNodeIndexerURL: "https://api-testnet.hyperion.xyz/v1/graphql/",
		HyperionAPIHost:            "https://api-testnet.hyperion.xyz/",
		AptosAPIKey:                "key",
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if sdk.Options.HyperionFullNodeIndexerURL != "https://api-testnet.hyperion.xyz" {
		t.Fatalf("indexer URL = %q", sdk.Options.HyperionFullNodeIndexerURL)
	}
	if sdk.Options.HyperionAPIHost != "https://api-testnet.hyperion.xyz" {
		t.Fatalf("api host = %q", sdk.Options.HyperionAPIHost)
	}
}

func TestNewRejectsMissingRequiredOptions(t *testing.T) {
	t.Parallel()

	if _, err := New(Options{}); err == nil {
		t.Fatal("New returned nil error for missing required options")
	}
}
