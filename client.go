package hyperion

import (
	"errors"
	"net/http"
)

// Network identifies the Aptos network used by the SDK.
type Network string

const (
	// NetworkMainnet selects Hyperion's Aptos mainnet deployment.
	NetworkMainnet Network = "mainnet"
	// NetworkTestnet selects Hyperion's Aptos testnet deployment.
	NetworkTestnet Network = "testnet"
)

const (
	// MainnetContractAddress is the Hyperion router contract address used by the upstream TypeScript SDK.
	MainnetContractAddress = "0x8b4a2c4bb53857c718a04c020b98f8c2e1f99a68b0f57389a8bf5434cd22e05c"
	// TestnetContractAddress is the Hyperion testnet router contract address used by the upstream TypeScript SDK.
	TestnetContractAddress = "0x69faed94da99abb7316cb3ec2eeaa1b961a47349fad8c584f67a930b0d14fec7"
)

// Options configures a Hyperion SDK client.
type Options struct {
	Network                    Network
	ContractAddress            string
	HyperionFullNodeIndexerURL string
	HyperionAPIHost            string
	AptosFullNodeURL           string
	AptosAPIKey                string
	HTTPClient                 *http.Client
	ViewExecutor               ViewExecutor
}

// InitOptions configures Init with Hyperion's built-in network defaults.
type InitOptions struct {
	Network          Network
	AptosFullNodeURL string
	AptosAPIKey      string
	HTTPClient       *http.Client
	ViewExecutor     ViewExecutor
}

// Client is the root Hyperion SDK handle. Services mirror the upstream TypeScript
// SDK modules while using Go contexts and typed request/response values.
type Client struct {
	Options     Options
	Request     *RequestClient
	APIRequest  *RequestClient
	ViewClient  ViewExecutor
	Pool        *PoolService
	Position    *PositionService
	Reward      *RewardService
	Swap        *SwapService
	PriceHub    *PriceHubService
	RateLimiter *RateLimiterService
	CoinWrapper *CoinWrapperService
}

// PoolService provides pool read and transaction-payload helpers.
type PoolService struct {
	client *Client
}

// PositionService provides position read and transaction-payload helpers.
type PositionService struct {
	client *Client
}

// RewardService provides reward read and transaction-payload helpers.
type RewardService struct {
	client *Client
}

// SwapService provides swap quote and transaction-payload helpers.
type SwapService struct {
	client *Client
}

// PriceHubService provides price_hub read-only view helpers.
type PriceHubService struct {
	client *Client
}

// RateLimiterService provides protocol guard and rate-limiter read-only view helpers.
type RateLimiterService struct {
	client *Client
}

// CoinWrapperService provides coin_wrapper identity read-only view helpers.
type CoinWrapperService struct {
	client *Client
}

// Init creates a Client using Hyperion's built-in mainnet or testnet defaults.
func Init(options InitOptions) (*Client, error) {
	switch options.Network {
	case NetworkMainnet:
		return New(Options{
			Network:                    NetworkMainnet,
			ContractAddress:            MainnetContractAddress,
			HyperionFullNodeIndexerURL: "https://api.hyperion.xyz",
			HyperionAPIHost:            "https://api.hyperion.xyz",
			AptosFullNodeURL:           options.AptosFullNodeURL,
			AptosAPIKey:                options.AptosAPIKey,
			HTTPClient:                 options.HTTPClient,
			ViewExecutor:               options.ViewExecutor,
		})
	case NetworkTestnet:
		return New(Options{
			Network:                    NetworkTestnet,
			ContractAddress:            TestnetContractAddress,
			HyperionFullNodeIndexerURL: "https://api-testnet.hyperion.xyz",
			HyperionAPIHost:            "https://api-testnet.hyperion.xyz",
			AptosFullNodeURL:           options.AptosFullNodeURL,
			AptosAPIKey:                options.AptosAPIKey,
			HTTPClient:                 options.HTTPClient,
			ViewExecutor:               options.ViewExecutor,
		})
	default:
		return nil, errors.New("network must be mainnet or testnet")
	}
}

// New creates a Client from explicit options.
func New(options Options) (*Client, error) {
	if options.Network == "" {
		return nil, errors.New("network is required")
	}
	if options.ContractAddress == "" {
		return nil, errors.New("contract address is required")
	}
	if options.HyperionFullNodeIndexerURL == "" {
		return nil, errors.New("hyperion full node indexer URL is required")
	}
	if options.HyperionAPIHost == "" {
		return nil, errors.New("hyperion API host is required")
	}

	options.HyperionFullNodeIndexerURL = normalizeAPIHost(options.HyperionFullNodeIndexerURL)
	options.HyperionAPIHost = normalizeAPIHost(options.HyperionAPIHost)
	if options.AptosFullNodeURL != "" {
		options.AptosFullNodeURL = normalizeAPIHost(options.AptosFullNodeURL)
	}
	viewExecutor := options.ViewExecutor
	if viewExecutor == nil && options.AptosFullNodeURL != "" {
		var err error
		viewExecutor, err = NewAptosViewExecutor(options.AptosFullNodeURL, options.AptosAPIKey, options.HTTPClient)
		if err != nil {
			return nil, err
		}
	}
	request := NewRequestClient(options.HyperionFullNodeIndexerURL, options.HTTPClient)
	apiRequest := NewRequestClient(options.HyperionAPIHost, options.HTTPClient)

	client := &Client{
		Options:    options,
		Request:    request,
		APIRequest: apiRequest,
		ViewClient: viewExecutor,
	}
	client.Pool = &PoolService{client: client}
	client.Position = &PositionService{client: client}
	client.Reward = &RewardService{client: client}
	client.Swap = &SwapService{client: client}
	client.PriceHub = &PriceHubService{client: client}
	client.RateLimiter = &RateLimiterService{client: client}
	client.CoinWrapper = &CoinWrapperService{client: client}

	return client, nil
}
