package hyperion

import "testing"

func TestParseIntegrationConfigDefaultsToDisabledMainnet(t *testing.T) {
	t.Parallel()

	cfg, err := parseIntegrationConfig(func(string) string { return "" })
	if err != nil {
		t.Fatalf("parseIntegrationConfig returned error: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("integration config enabled without HYPERION_INTEGRATION=1")
	}
	if cfg.Network != NetworkMainnet {
		t.Fatalf("network = %q, want %q", cfg.Network, NetworkMainnet)
	}
	if cfg.SwapAmount != "1000" {
		t.Fatalf("swap amount = %q, want default 1000", cfg.SwapAmount)
	}
	if cfg.AggregateAmount != "1000" {
		t.Fatalf("aggregate amount = %q, want default 1000", cfg.AggregateAmount)
	}
}

func TestParseIntegrationConfigReadsEnvironment(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"HYPERION_INTEGRATION": "1",
		"HYPERION_NETWORK":     "testnet",
		"HYPERION_API_HOST":    "https://hyperion.example/v1/graphql/",
		"APTOS_FULLNODE_URL":   "https://aptos.example/v1/",
		"APTOS_API_KEY":        "aptos-key",
		"HYPERION_SWAP_FROM":   "0x1",
		"HYPERION_SWAP_TO":     "0x2",
		"HYPERION_SWAP_AMOUNT": "42",
		"HYPERION_AGG_FROM":    "0x3",
		"HYPERION_AGG_INPUT":   "0x4",
		"HYPERION_AGG_TO":      "0x5",
		"HYPERION_AGG_AMOUNT":  "84",
	}

	cfg, err := parseIntegrationConfig(func(key string) string {
		return env[key]
	})
	if err != nil {
		t.Fatalf("parseIntegrationConfig returned error: %v", err)
	}
	if !cfg.Enabled {
		t.Fatal("integration config is disabled")
	}
	if cfg.Network != NetworkTestnet {
		t.Fatalf("network = %q, want %q", cfg.Network, NetworkTestnet)
	}
	if cfg.HyperionAPIHost != "https://hyperion.example/v1/graphql/" {
		t.Fatalf("hyperion api host = %q", cfg.HyperionAPIHost)
	}
	if cfg.AptosFullNodeURL != "https://aptos.example/v1/" {
		t.Fatalf("aptos fullnode URL = %q", cfg.AptosFullNodeURL)
	}
	if cfg.AptosAPIKey != "aptos-key" {
		t.Fatalf("aptos API key = %q", cfg.AptosAPIKey)
	}
	if cfg.SwapFrom != "0x1" || cfg.SwapTo != "0x2" || cfg.SwapAmount != "42" {
		t.Fatalf("swap config = %#v", cfg)
	}
	if cfg.AggregateFrom != "0x3" || cfg.AggregateInput != "0x4" || cfg.AggregateTo != "0x5" || cfg.AggregateAmount != "84" {
		t.Fatalf("aggregate config = %#v", cfg)
	}
}

func TestParseIntegrationConfigIgnoresUnknownNetworkWhenDisabled(t *testing.T) {
	t.Parallel()

	cfg, err := parseIntegrationConfig(func(key string) string {
		if key == "HYPERION_NETWORK" {
			return "devnet"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("parseIntegrationConfig returned error while integration was disabled: %v", err)
	}
	if cfg.Network != NetworkMainnet {
		t.Fatalf("network = %q, want %q", cfg.Network, NetworkMainnet)
	}
}

func TestParseIntegrationConfigRejectsUnknownNetworkWhenEnabled(t *testing.T) {
	t.Parallel()

	_, err := parseIntegrationConfig(func(key string) string {
		switch key {
		case "HYPERION_INTEGRATION":
			return "1"
		case "HYPERION_NETWORK":
			return "devnet"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatal("parseIntegrationConfig accepted an unknown network while integration was enabled")
	}
}

func TestIntegrationClientOptionsUseDefaultsAndOverrides(t *testing.T) {
	t.Parallel()

	mainnet := integrationConfig{
		Network:          NetworkMainnet,
		AptosFullNodeURL: "https://aptos.example/v1",
		AptosAPIKey:      "key",
	}
	mainnetOptions := mainnet.clientOptions()
	if mainnetOptions.ContractAddress != MainnetContractAddress {
		t.Fatalf("mainnet contract = %q", mainnetOptions.ContractAddress)
	}
	if mainnetOptions.HyperionAPIHost != "https://api.hyperion.xyz" {
		t.Fatalf("mainnet API host = %q", mainnetOptions.HyperionAPIHost)
	}
	if mainnetOptions.AptosFullNodeURL != "https://aptos.example/v1" {
		t.Fatalf("aptos fullnode URL = %q", mainnetOptions.AptosFullNodeURL)
	}

	testnet := integrationConfig{
		Network:          NetworkTestnet,
		HyperionAPIHost:  "https://custom.example",
		AptosFullNodeURL: "https://aptos.example",
	}
	testnetOptions := testnet.clientOptions()
	if testnetOptions.ContractAddress != TestnetContractAddress {
		t.Fatalf("testnet contract = %q", testnetOptions.ContractAddress)
	}
	if testnetOptions.HyperionFullNodeIndexerURL != "https://custom.example" {
		t.Fatalf("testnet indexer URL = %q", testnetOptions.HyperionFullNodeIndexerURL)
	}
	if testnetOptions.HyperionAPIHost != "https://custom.example" {
		t.Fatalf("testnet API host = %q", testnetOptions.HyperionAPIHost)
	}
}
