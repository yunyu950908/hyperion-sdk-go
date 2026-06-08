package hyperion

import (
	"context"
	"errors"
	"fmt"
)

// ErrAggregateSubmitAdapterUnsupported marks the current gap between Hyperion's
// aggregate recorder plan and a submit-ready Aptos transaction. Current Aptos Go
// SDK releases can wrap compiled script bytecode, but do not expose the
// TypeScript Dynamic Script Composer compiler needed to turn batched call/result
// references into that bytecode.
var ErrAggregateSubmitAdapterUnsupported = errors.New("aggregate submit-ready transaction adapter is unsupported")

const defaultAggregateSubmitAdapterUnsupportedReason = "current Aptos Go SDK releases can build ScriptPayload values when compiled bytecode is already available, but do not expose a TypeScript Dynamic Script Composer-compatible compiler for batched call/result references"

// BuildAggregateSwapSubmitTransactionArgs configures aggregate submit adapter
// execution. The SDK still does not sign, simulate, or submit transactions; the
// adapter is responsible for turning the deterministic call plan into whatever
// transaction object or bytes a downstream Aptos transaction layer requires.
type BuildAggregateSwapSubmitTransactionArgs struct {
	Route         AggregateSwapInfoResult
	PartnershipID string
	Adapter       AggregateSwapSubmitAdapter
}

// BuildAggregateSwapSubmitPlanArgs configures read-only aggregate submit plan
// generation. The returned plan is not a wallet payload; callers still need a
// transaction composer before a wallet can sign or submit an aggregate swap.
type BuildAggregateSwapSubmitPlanArgs struct {
	Route         AggregateSwapInfoResult
	PartnershipID string
}

// AggregateSwapSubmitAdapter converts a deterministic aggregate call plan into
// adapter-specific submit-ready transaction data.
type AggregateSwapSubmitAdapter interface {
	BuildAggregateSwapSubmitTransaction(context.Context, AggregateSwapSubmitPlan) (*AggregateSwapSubmitTransaction, error)
}

// AggregateSwapSubmitPlan is the deterministic call plan passed to a submit
// adapter. Calls mirror AggregateSwapRecorder output and are safe for tests,
// audits, external TS/WASM bridge adapters, or future native Go adapters.
type AggregateSwapSubmitPlan struct {
	Route             AggregateSwapInfoResult     `json:"route"`
	PartnershipID     string                      `json:"partnershipId,omitempty"`
	ExactIn           bool                        `json:"exactIn"`
	RouteSplits       int                         `json:"routeSplits"`
	RefundRouteSplits int                         `json:"refundRouteSplits"`
	Calls             []AggregateSwapComposerCall `json:"calls"`
}

// AggregateSwapSubmitTransaction is adapter-specific transaction output. A
// native Go adapter can attach raw BCS bytes, a wallet adapter can attach a
// signing message, and Raw can carry an SDK-specific transaction object.
type AggregateSwapSubmitTransaction struct {
	PayloadType    string `json:"payloadType,omitempty"`
	PayloadBytes   []byte `json:"payloadBytes,omitempty"`
	SigningMessage []byte `json:"signingMessage,omitempty"`
	Raw            any    `json:"-"`
}

// UnsupportedAggregateSwapSubmitAdapter is the built-in adapter for current
// upstream gaps. It gives callers a structured, testable error instead of
// treating recorder output as a submit-ready transaction.
type UnsupportedAggregateSwapSubmitAdapter struct {
	Reason string
}

// NewUnsupportedAggregateSwapSubmitAdapter returns an adapter that always
// reports the current Dynamic Script Composer gap in Aptos Go SDK support.
func NewUnsupportedAggregateSwapSubmitAdapter() UnsupportedAggregateSwapSubmitAdapter {
	return UnsupportedAggregateSwapSubmitAdapter{}
}

// BuildAggregateSwapSubmitTransaction reports the unsupported upstream gap.
func (a UnsupportedAggregateSwapSubmitAdapter) BuildAggregateSwapSubmitTransaction(_ context.Context, plan AggregateSwapSubmitPlan) (*AggregateSwapSubmitTransaction, error) {
	reason := a.Reason
	if reason == "" {
		reason = defaultAggregateSubmitAdapterUnsupportedReason
	}
	return nil, &AggregateSwapSubmitAdapterUnsupportedError{
		Reason: reason,
		Calls:  len(plan.Calls),
	}
}

// AggregateSwapSubmitAdapterUnsupportedError describes why a submit-ready
// aggregate transaction could not be built.
type AggregateSwapSubmitAdapterUnsupportedError struct {
	Reason string
	Calls  int
}

func (e *AggregateSwapSubmitAdapterUnsupportedError) Error() string {
	if e == nil {
		return ErrAggregateSubmitAdapterUnsupported.Error()
	}
	return fmt.Sprintf("%s: %s (composer calls: %d)", ErrAggregateSubmitAdapterUnsupported, e.Reason, e.Calls)
}

func (e *AggregateSwapSubmitAdapterUnsupportedError) Unwrap() error {
	return ErrAggregateSubmitAdapterUnsupported
}

// BuildAggregateSwapSubmitTransaction composes the route into a deterministic
// aggregate call plan and delegates submit-ready transaction construction to the
// provided adapter.
func (s *SwapService) BuildAggregateSwapSubmitTransaction(ctx context.Context, args BuildAggregateSwapSubmitTransactionArgs) (*AggregateSwapSubmitTransaction, error) {
	if args.Adapter == nil {
		return nil, errors.New("aggregate submit adapter is required")
	}

	plan, err := s.BuildAggregateSwapSubmitPlan(BuildAggregateSwapSubmitPlanArgs{
		Route:         args.Route,
		PartnershipID: args.PartnershipID,
	})
	if err != nil {
		return nil, err
	}
	result, err := args.Adapter.BuildAggregateSwapSubmitTransaction(ctx, plan)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("aggregate submit adapter returned nil transaction")
	}
	return result, nil
}

// BuildAggregateSwapSubmitPlan composes an aggregate route into a deterministic,
// read-only call plan. The plan is useful for audits, UI handoff, or external
// wallet/composer integrations, but it is not a submit-ready Aptos transaction.
func (s *SwapService) BuildAggregateSwapSubmitPlan(args BuildAggregateSwapSubmitPlanArgs) (AggregateSwapSubmitPlan, error) {
	recorder := NewAggregateSwapRecorder()
	if err := s.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:         args.Route,
		Composer:      recorder,
		PartnershipID: args.PartnershipID,
	}); err != nil {
		return AggregateSwapSubmitPlan{}, err
	}

	plan := AggregateSwapSubmitPlan{
		Route:             cloneAggregateSwapInfoResult(args.Route),
		PartnershipID:     args.PartnershipID,
		ExactIn:           args.Route.ExactIn,
		RouteSplits:       len(args.Route.Quotes.Route),
		RefundRouteSplits: len(args.Route.Quotes.RefundRoute),
		Calls:             cloneAggregateSwapComposerCalls(recorder.Calls),
	}
	return plan, nil
}

func cloneAggregateSwapComposerCalls(calls []AggregateSwapComposerCall) []AggregateSwapComposerCall {
	out := make([]AggregateSwapComposerCall, len(calls))
	for i, call := range calls {
		out[i] = call
		out[i].TypeArguments = append([]string(nil), call.TypeArguments...)
		out[i].FunctionArguments = cloneAggregateSwapCallArguments(call.FunctionArguments)
	}
	return out
}

func cloneAggregateSwapInfoResult(route AggregateSwapInfoResult) AggregateSwapInfoResult {
	out := route
	out.Quotes.Route = cloneAggregateSwapRoutes(route.Quotes.Route)
	out.Quotes.RefundRoute = cloneAggregateSwapRoutes(route.Quotes.RefundRoute)
	return out
}

func cloneAggregateSwapRoutes(routes []AggregateSwapRoute) []AggregateSwapRoute {
	out := make([]AggregateSwapRoute, len(routes))
	for i, route := range routes {
		out[i] = route
		out[i].RouteTaken = append([]AggregateSwapRouteTaken(nil), route.RouteTaken...)
	}
	return out
}

func cloneAggregateSwapCallArguments(args []AggregateSwapCallArgument) []AggregateSwapCallArgument {
	out := make([]AggregateSwapCallArgument, len(args))
	for i, arg := range args {
		out[i] = arg
		if arg.Result != nil {
			result := *arg.Result
			out[i].Result = &result
		}
	}
	return out
}
