package hyperion

import (
	"errors"
	"fmt"
)

const (
	metadataType          = "0x1::fungible_asset::Metadata"
	cellanaContract       = "0x4bf51972879e3b95c4781a5cdcb9e1ee24ef483e7d22f2d903626f126df62bd1"
	thalaSwapV2Contract   = "0x7730cd28ee1cdc9e999336cbc430f99e7c44397c0aa77516f6f23a78559bb5"
	aggregateArgLiteral   = "literal"
	aggregateArgResult    = "result"
	aggregateArgSigner    = "signer"
	aggregateArgCopy      = "copy"
	aggregateArgBorrow    = "borrow"
	aggregateArgBorrowMut = "borrow_mut"
)

// GenerateAggregateSwapTransactionScriptArgs configures aggregate swap script
// composition. The composer is an adapter boundary: the built-in recorder stores
// the batched calls deterministically, while a future Aptos SDK adapter can turn
// the same calls into a submit-ready transaction.
type GenerateAggregateSwapTransactionScriptArgs struct {
	Route         AggregateSwapInfoResult
	Composer      AggregateSwapComposer
	PartnershipID string
}

// AggregateSwapComposer records or executes script-composer batched calls.
type AggregateSwapComposer interface {
	AddBatchedCall(AggregateSwapComposerCall) ([]AggregateSwapCallArgument, error)
}

// AggregateSwapComposerCall mirrors one TypeScript AptosScriptComposer
// addBatchedCalls invocation.
type AggregateSwapComposerCall struct {
	Function          string                      `json:"function"`
	TypeArguments     []string                    `json:"typeArguments"`
	FunctionArguments []AggregateSwapCallArgument `json:"functionArguments"`
	ReturnCount       int                         `json:"returnCount"`
}

// AggregateSwapCallArgument is a symbolic call argument used by the recorder
// and composer adapters.
type AggregateSwapCallArgument struct {
	Kind     string                       `json:"kind"`
	Value    any                          `json:"value,omitempty"`
	Result   *AggregateSwapResultArgument `json:"result,omitempty"`
	Modifier string                       `json:"modifier,omitempty"`
}

// AggregateSwapResultArgument identifies one return value from a previous
// batched call.
type AggregateSwapResultArgument struct {
	Call  int `json:"call"`
	Index int `json:"index"`
}

// Copy mirrors CallArgument.copy in the TypeScript composer SDK.
func (a AggregateSwapCallArgument) Copy() AggregateSwapCallArgument {
	a.Modifier = aggregateArgCopy
	return a
}

// Borrow mirrors CallArgument.borrow in the TypeScript composer SDK.
func (a AggregateSwapCallArgument) Borrow() AggregateSwapCallArgument {
	a.Modifier = aggregateArgBorrow
	return a
}

// BorrowMut mirrors CallArgument.borrowMut in the TypeScript composer SDK.
func (a AggregateSwapCallArgument) BorrowMut() AggregateSwapCallArgument {
	a.Modifier = aggregateArgBorrowMut
	return a
}

// AggregateSwapRecorder is a deterministic in-memory composer useful for tests,
// audits, and future Aptos SDK adapter development.
type AggregateSwapRecorder struct {
	Calls []AggregateSwapComposerCall `json:"calls"`
}

// NewAggregateSwapRecorder creates an empty aggregate swap composer recorder.
func NewAggregateSwapRecorder() *AggregateSwapRecorder {
	return &AggregateSwapRecorder{}
}

// AddBatchedCall appends a call and returns symbolic references to its return
// values.
func (r *AggregateSwapRecorder) AddBatchedCall(call AggregateSwapComposerCall) ([]AggregateSwapCallArgument, error) {
	if r == nil {
		return nil, errors.New("aggregate swap recorder is nil")
	}
	if call.ReturnCount < 0 {
		return nil, errors.New("return count can not be negative")
	}
	call.TypeArguments = append([]string(nil), call.TypeArguments...)
	call.FunctionArguments = append([]AggregateSwapCallArgument(nil), call.FunctionArguments...)

	callIndex := len(r.Calls)
	r.Calls = append(r.Calls, call)
	results := make([]AggregateSwapCallArgument, call.ReturnCount)
	for i := range results {
		results[i] = resultArg(callIndex, i)
	}
	return results, nil
}

// GenerateAggregateSwapTransactionScript composes aggregate swap batched calls
// into the provided composer. It mirrors the upstream TypeScript helper's call
// order but intentionally does not serialize a submit-ready transaction itself.
func (s *SwapService) GenerateAggregateSwapTransactionScript(args GenerateAggregateSwapTransactionScriptArgs) error {
	if s.client.Options.Network != NetworkMainnet {
		return errors.New("aggregate swap is only supported on MAINNET")
	}
	if args.Composer == nil {
		return errors.New("aggregate swap composer is required")
	}

	sender, err := getSender(args.Composer)
	if err != nil {
		return err
	}

	amountIn := args.Route.FromTokenAmount
	if !args.Route.ExactIn {
		amountIn = args.Route.MaxFromTokenAmount
	}

	metadata, err := addressToObject(args.Composer, metadataType, args.Route.FromToken.Address)
	if err != nil {
		return err
	}
	tokenIn, err := withdrawFA(args.Composer, metadata, amountIn)
	if err != nil {
		return err
	}
	if _, err := amountFA(args.Composer, tokenIn); err != nil {
		return err
	}

	tokenOut, err := zeroFA(args.Composer, args.Route.ToToken.Address)
	if err != nil {
		return err
	}
	for _, route := range args.Route.Quotes.Route {
		routeTokenIn, err := extractFA(args.Composer, tokenIn, route.AmountIn)
		if err != nil {
			return err
		}
		ret, err := s.bigRouteCompose(args.Composer, route.RouteTaken, routeTokenIn, sender.Copy(), args.PartnershipID)
		if err != nil {
			return err
		}
		if err := mergeFA(args.Composer, tokenOut, ret); err != nil {
			return err
		}
	}

	if args.Route.ExactIn {
		if err := faAmountCheck(args.Composer, tokenOut, args.Route.MinToTokenAmount); err != nil {
			return err
		}
		if err := depositFA(args.Composer, sender.Copy(), tokenOut); err != nil {
			return err
		}
		return depositFA(args.Composer, sender.Copy(), tokenIn)
	}

	if err := faAmountCheck(args.Composer, tokenOut, args.Route.ToTokenAmount); err != nil {
		return err
	}
	if err := depositExactFA(args.Composer, sender.Copy(), tokenOut, args.Route.ToTokenAmount); err != nil {
		return err
	}

	tokenInRefund, err := zeroFA(args.Composer, args.Route.FromToken.Address)
	if err != nil {
		return err
	}
	amountInArg, err := amountFA(args.Composer, tokenOut)
	if err != nil {
		return err
	}

	percentage := 0
	for _, route := range args.Route.Quotes.RefundRoute {
		percentageIn := route.Percentage
		percentage += percentageIn
		if percentage == 100 {
			percentageIn = 100
		}
		routeTokenIn, err := faSplitProportionally(args.Composer, tokenOut, amountInArg.Copy(), percentageIn)
		if err != nil {
			return err
		}
		ret, err := s.bigRouteCompose(args.Composer, route.RouteTaken, routeTokenIn, sender.Copy(), args.PartnershipID)
		if err != nil {
			return err
		}
		if err := mergeFA(args.Composer, tokenInRefund, ret); err != nil {
			return err
		}
	}
	if err := depositFA(args.Composer, sender.Copy(), tokenOut); err != nil {
		return err
	}
	if err := depositFA(args.Composer, sender.Copy(), tokenIn); err != nil {
		return err
	}
	if err := depositFA(args.Composer, sender.Copy(), tokenInRefund); err != nil {
		return err
	}
	if percentage != 100 {
		return errors.New("not 100% refund")
	}
	return nil
}

func (s *SwapService) bigRouteCompose(composer AggregateSwapComposer, routes []AggregateSwapRouteTaken, tokenIn AggregateSwapCallArgument, sender AggregateSwapCallArgument, partnershipID string) (AggregateSwapCallArgument, error) {
	for _, singlePair := range routes {
		tokenOut, err := zeroFA(composer, singlePair.ToToken.TokenType)
		if err != nil {
			return AggregateSwapCallArgument{}, err
		}
		ret, err := s.composeAggregateRoute(composer, singlePair, tokenIn, sender.Copy(), partnershipID)
		if err != nil {
			return AggregateSwapCallArgument{}, err
		}
		if err := mergeFA(composer, tokenOut, ret); err != nil {
			return AggregateSwapCallArgument{}, err
		}
		tokenIn = tokenOut
	}
	return tokenIn, nil
}

func (s *SwapService) composeAggregateRoute(composer AggregateSwapComposer, route AggregateSwapRouteTaken, tokenIn AggregateSwapCallArgument, sender AggregateSwapCallArgument, partnershipID string) (AggregateSwapCallArgument, error) {
	switch route.DexName {
	case "Hyperion":
		return s.hyperionSwapFAToFA(composer, sender.Copy(), route, tokenIn, partnershipID)
	case "Cellana":
		stable, err := cellanaStableFlag(route.PoolType)
		if err != nil {
			return AggregateSwapCallArgument{}, err
		}
		return cellanaSwap(composer, tokenIn, route.ToToken.TokenType, stable)
	case "ThalaSwapV2":
		return thalaSwapExactIn(composer, route, tokenIn)
	case "EmojiCoin":
		return emojiSwap(composer, route, tokenIn)
	default:
		return AggregateSwapCallArgument{}, fmt.Errorf("DEX not supported: %s", route.DexName)
	}
}

func (s *SwapService) hyperionSwapFAToFA(composer AggregateSwapComposer, sender AggregateSwapCallArgument, route AggregateSwapRouteTaken, faIn AggregateSwapCallArgument, partnershipID string) (AggregateSwapCallArgument, error) {
	amount, err := amountFA(composer, faIn)
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	if partnershipID == "" {
		partnershipID = AggregatorPartnerName
	}
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      s.client.Options.ContractAddress + "::partnership::swap",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			literalArg(route.PoolID),
			literalArg(route.A2B),
			literalArg(true),
			amount,
			faIn,
			literalArg(route.SqrtPriceLimit),
			literalArg(partnershipID),
		},
		ReturnCount: 3,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	if err := depositFA(composer, sender, ret[1]); err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[2], nil
}

func zeroFA(composer AggregateSwapComposer, assetTypeString string) (AggregateSwapCallArgument, error) {
	assetMeta, err := addressToObject(composer, metadataType, assetTypeString)
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:          "0x1::fungible_asset::zero",
		TypeArguments:     []string{metadataType},
		FunctionArguments: []AggregateSwapCallArgument{assetMeta},
		ReturnCount:       1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func extractFA(composer AggregateSwapComposer, assetToSplit AggregateSwapCallArgument, amount string) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      "0x1::fungible_asset::extract",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			assetToSplit.BorrowMut(),
			literalArg(amount),
		},
		ReturnCount: 1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func mergeFA(composer AggregateSwapComposer, asset AggregateSwapCallArgument, assetToMerge AggregateSwapCallArgument) error {
	_, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      "0x1::fungible_asset::merge",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			asset.BorrowMut(),
			assetToMerge,
		},
		ReturnCount: 0,
	})
	return err
}

func withdrawFA(composer AggregateSwapComposer, metadata AggregateSwapCallArgument, amount string) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      "0x1::primary_fungible_store::withdraw",
		TypeArguments: []string{metadataType},
		FunctionArguments: []AggregateSwapCallArgument{
			signerArg(0),
			metadata,
			literalArg(amount),
		},
		ReturnCount: 1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func depositFA(composer AggregateSwapComposer, receiver AggregateSwapCallArgument, fa AggregateSwapCallArgument) error {
	_, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      "0x1::primary_fungible_store::deposit",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			receiver,
			fa,
		},
		ReturnCount: 0,
	})
	return err
}

func amountFA(composer AggregateSwapComposer, fa AggregateSwapCallArgument) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:          "0x1::fungible_asset::amount",
		TypeArguments:     []string{},
		FunctionArguments: []AggregateSwapCallArgument{fa.Borrow()},
		ReturnCount:       1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func addressToObject(composer AggregateSwapComposer, objectType, objectAddress string) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      "0x1::object::address_to_object",
		TypeArguments: []string{objectType},
		FunctionArguments: []AggregateSwapCallArgument{
			literalArg(objectAddress),
		},
		ReturnCount: 1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func faSplitProportionally(composer AggregateSwapComposer, fa AggregateSwapCallArgument, amountIn AggregateSwapCallArgument, percent int) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      AggregateToolContractAddress + "::tool::split_fa_proportionlly",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			fa.BorrowMut(),
			amountIn,
			literalArg(percent * 100),
		},
		ReturnCount: 1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func faAmountCheck(composer AggregateSwapComposer, fa AggregateSwapCallArgument, amount string) error {
	_, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      AggregateToolContractAddress + "::tool::fa_amount_check",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			fa.Borrow(),
			literalArg(amount),
		},
		ReturnCount: 0,
	})
	return err
}

func depositExactFA(composer AggregateSwapComposer, receiver AggregateSwapCallArgument, fa AggregateSwapCallArgument, amount string) error {
	_, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      AggregateToolContractAddress + "::tool::deposit_fa_exact",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			receiver,
			fa.BorrowMut(),
			literalArg(amount),
		},
		ReturnCount: 0,
	})
	return err
}

func getSender(composer AggregateSwapComposer) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      AggregateToolContractAddress + "::tool::get_signer_address",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			signerArg(0),
		},
		ReturnCount: 1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func cellanaSwap(composer AggregateSwapComposer, assetIn AggregateSwapCallArgument, outMetaID string, stable bool) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      cellanaContract + "::router::swap",
		TypeArguments: []string{},
		FunctionArguments: []AggregateSwapCallArgument{
			assetIn,
			literalArg(0),
			literalArg(outMetaID),
			literalArg(stable),
		},
		ReturnCount: 1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func cellanaStableFlag(poolType string) (bool, error) {
	switch poolType {
	case "stable":
		return true, nil
	case "unstable":
		return false, nil
	default:
		return false, errors.New("pool type mismatch")
	}
}

func thalaSwapExactIn(composer AggregateSwapComposer, route AggregateSwapRouteTaken, assetIn AggregateSwapCallArgument) (AggregateSwapCallArgument, error) {
	switch route.PoolType {
	case "stable":
		outMeta, err := addressToObject(composer, metadataType, route.ToToken.TokenType)
		if err != nil {
			return AggregateSwapCallArgument{}, err
		}
		return thalaSwapCall(composer, "swap_exact_in_stable", []AggregateSwapCallArgument{
			signerArg(0),
			literalArg(route.PoolID),
			assetIn,
			outMeta,
		})
	case "weighted":
		return thalaSwapCall(composer, "swap_exact_in_weighted", []AggregateSwapCallArgument{
			signerArg(0),
			literalArg(route.PoolID),
			assetIn,
			literalArg(route.ToToken.TokenType),
		})
	case "meta":
		outMeta, err := addressToObject(composer, metadataType, route.ToToken.TokenType)
		if err != nil {
			return AggregateSwapCallArgument{}, err
		}
		return thalaSwapCall(composer, "swap_exact_in_metastable", []AggregateSwapCallArgument{
			signerArg(0),
			literalArg(route.PoolID),
			assetIn,
			outMeta,
		})
	default:
		return AggregateSwapCallArgument{}, errors.New("type mismatch")
	}
}

func thalaSwapCall(composer AggregateSwapComposer, function string, args []AggregateSwapCallArgument) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:          thalaSwapV2Contract + "::pool::" + function,
		TypeArguments:     []string{},
		FunctionArguments: args,
		ReturnCount:       1,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[0], nil
}

func emojiSwap(composer AggregateSwapComposer, route AggregateSwapRouteTaken, tokenInput AggregateSwapCallArgument) (AggregateSwapCallArgument, error) {
	ret, err := addBatchedCall(composer, AggregateSwapComposerCall{
		Function:      AggregateToolContractAddress + "::tool::swap_in_emoji",
		TypeArguments: []string{route.FirstType, route.SecondType},
		FunctionArguments: []AggregateSwapCallArgument{
			signerArg(0),
			literalArg(route.PoolID),
			literalArg(route.IsSell),
			literalArg(route.Integrator),
			literalArg(route.IntegratorFee),
			tokenInput,
		},
		ReturnCount: 2,
	})
	if err != nil {
		return AggregateSwapCallArgument{}, err
	}
	return ret[1], nil
}

func addBatchedCall(composer AggregateSwapComposer, call AggregateSwapComposerCall) ([]AggregateSwapCallArgument, error) {
	if composer == nil {
		return nil, errors.New("aggregate swap composer is required")
	}
	return composer.AddBatchedCall(call)
}

func literalArg(value any) AggregateSwapCallArgument {
	return AggregateSwapCallArgument{Kind: aggregateArgLiteral, Value: value}
}

func signerArg(index int) AggregateSwapCallArgument {
	return AggregateSwapCallArgument{Kind: aggregateArgSigner, Value: index}
}

func resultArg(call, index int) AggregateSwapCallArgument {
	return AggregateSwapCallArgument{
		Kind: aggregateArgResult,
		Result: &AggregateSwapResultArgument{
			Call:  call,
			Index: index,
		},
	}
}

func aggregateCallArgumentEqual(a, b AggregateSwapCallArgument) bool {
	if a.Kind != b.Kind || a.Modifier != b.Modifier {
		return false
	}
	if a.Value != b.Value {
		return false
	}
	if a.Result == nil || b.Result == nil {
		return a.Result == nil && b.Result == nil
	}
	return a.Result.Call == b.Result.Call && a.Result.Index == b.Result.Index
}
