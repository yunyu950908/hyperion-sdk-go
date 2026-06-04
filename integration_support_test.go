package hyperion

import (
	"errors"
	"os"
	"strings"
	"testing"
)

type integrationConfig struct {
	Enabled          bool
	Network          Network
	HyperionAPIHost  string
	AptosFullNodeURL string
	AptosAPIKey      string
	SwapFrom         string
	SwapTo           string
	SwapAmount       string
	AggregateFrom    string
	AggregateInput   string
	AggregateTo      string
	AggregateAmount  string
}

func parseIntegrationConfig(getenv func(string) string) (integrationConfig, error) {
	enabled := getenv("HYPERION_INTEGRATION") == "1"
	network := Network(strings.ToLower(strings.TrimSpace(getenv("HYPERION_NETWORK"))))
	if network == "" {
		network = NetworkMainnet
	}
	if network != NetworkMainnet && network != NetworkTestnet {
		if enabled {
			return integrationConfig{}, errors.New("HYPERION_NETWORK must be mainnet or testnet")
		}
		network = NetworkMainnet
	}

	cfg := integrationConfig{
		Enabled:          enabled,
		Network:          network,
		HyperionAPIHost:  getenv("HYPERION_API_HOST"),
		AptosFullNodeURL: getenv("APTOS_FULLNODE_URL"),
		AptosAPIKey:      getenv("APTOS_API_KEY"),
		SwapFrom:         getenv("HYPERION_SWAP_FROM"),
		SwapTo:           getenv("HYPERION_SWAP_TO"),
		SwapAmount:       envDefault(getenv, "HYPERION_SWAP_AMOUNT", "1000"),
		AggregateFrom:    getenv("HYPERION_AGG_FROM"),
		AggregateInput:   getenv("HYPERION_AGG_INPUT"),
		AggregateTo:      getenv("HYPERION_AGG_TO"),
		AggregateAmount:  envDefault(getenv, "HYPERION_AGG_AMOUNT", "1000"),
	}
	return cfg, nil
}

func envDefault(getenv func(string) string, key, fallback string) string {
	value := getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadIntegrationConfig(t *testing.T) integrationConfig {
	t.Helper()

	cfg, err := parseIntegrationConfig(os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled {
		t.Skip("set HYPERION_INTEGRATION=1 to run live integration tests")
	}
	return cfg
}

func (cfg integrationConfig) clientOptions() Options {
	contractAddress := MainnetContractAddress
	hyperionAPIHost := "https://api.hyperion.xyz"
	if cfg.Network == NetworkTestnet {
		contractAddress = TestnetContractAddress
		hyperionAPIHost = "https://api-testnet.hyperion.xyz"
	}
	if cfg.HyperionAPIHost != "" {
		hyperionAPIHost = cfg.HyperionAPIHost
	}

	return Options{
		Network:                    cfg.Network,
		ContractAddress:            contractAddress,
		HyperionFullNodeIndexerURL: hyperionAPIHost,
		HyperionAPIHost:            hyperionAPIHost,
		AptosFullNodeURL:           cfg.AptosFullNodeURL,
		AptosAPIKey:                cfg.AptosAPIKey,
	}
}

func newIntegrationClient(t *testing.T, cfg integrationConfig) *Client {
	t.Helper()

	sdk, err := New(cfg.clientOptions())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return sdk
}
