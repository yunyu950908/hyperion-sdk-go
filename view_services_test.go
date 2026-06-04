package hyperion

import (
	"context"
	"errors"
	"testing"
)

type capturingViewExecutor struct {
	payloads []EntryFunctionPayload
	values   []any
	err      error
}

func (e *capturingViewExecutor) View(ctx context.Context, payload EntryFunctionPayload) ([]any, error) {
	e.payloads = append(e.payloads, payload)
	if e.err != nil {
		return nil, e.err
	}
	return e.values, nil
}

func TestPoolViewMethodsUseExecutor(t *testing.T) {
	t.Parallel()

	executor := &capturingViewExecutor{values: []any{"321"}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	args := EstCurrencyAAmountFromBArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyBAmount:  "2000",
	}

	values, err := sdk.Pool.EstCurrencyAAmountFromB(context.Background(), args)
	if err != nil {
		t.Fatalf("EstCurrencyAAmountFromB returned error: %v", err)
	}
	expectedPayload, err := sdk.Pool.EstCurrencyAAmountFromBPayload(args)
	if err != nil {
		t.Fatalf("EstCurrencyAAmountFromBPayload returned error: %v", err)
	}

	assertArguments(t, values, []any{"321"})
	if len(executor.payloads) != 1 {
		t.Fatalf("captured %d payloads, want 1", len(executor.payloads))
	}
	if executor.payloads[0].Function != expectedPayload.Function {
		t.Fatalf("function = %q, want %q", executor.payloads[0].Function, expectedPayload.Function)
	}
	assertArguments(t, executor.payloads[0].FunctionArguments, expectedPayload.FunctionArguments)
}

func TestPoolViewMethodsPropagatePayloadBuilderErrors(t *testing.T) {
	t.Parallel()

	executor := &capturingViewExecutor{values: []any{"unused"}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	_, err := sdk.Pool.EstCurrencyBAmountFromA(context.Background(), EstCurrencyBAmountFromAArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "invalid",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyAAmount:  "1000",
	})
	if err == nil {
		t.Fatal("EstCurrencyBAmountFromA accepted an invalid tick")
	}
	if len(executor.payloads) != 0 {
		t.Fatalf("executor captured %d payloads, want 0", len(executor.payloads))
	}
}

func TestPositionAndRewardViewMethodsUseExecutor(t *testing.T) {
	t.Parallel()

	executor := &capturingViewExecutor{values: []any{"100", "200"}}
	sdk := newMainnetClientWithViewExecutor(t, executor)

	positionValues, err := sdk.Position.FetchTokensAmountByPositionID(context.Background(), "pos-1")
	if err != nil {
		t.Fatalf("FetchTokensAmountByPositionID returned error: %v", err)
	}
	rewardValues, err := sdk.Reward.FetchRewards(context.Background(), "pos-1")
	if err != nil {
		t.Fatalf("FetchRewards returned error: %v", err)
	}

	assertArguments(t, positionValues, []any{"100", "200"})
	assertArguments(t, rewardValues, []any{"100", "200"})
	if len(executor.payloads) != 2 {
		t.Fatalf("captured %d payloads, want 2", len(executor.payloads))
	}
	if executor.payloads[0].Function != sdk.Position.FetchTokensAmountByPositionIDPayload("pos-1").Function {
		t.Fatalf("position function = %q", executor.payloads[0].Function)
	}
	if executor.payloads[1].Function != sdk.Reward.FetchRewardsPayload("pos-1").Function {
		t.Fatalf("reward function = %q", executor.payloads[1].Function)
	}
}

func TestViewMethodsPropagateExecutorErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("view unavailable")
	sdk := newMainnetClientWithViewExecutor(t, &capturingViewExecutor{err: expectedErr})

	_, err := sdk.Position.FetchTokensAmountByPositionID(context.Background(), "pos-1")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}

func newMainnetClientWithViewExecutor(t *testing.T, executor ViewExecutor) *Client {
	t.Helper()

	sdk, err := New(Options{
		Network:                    NetworkMainnet,
		ContractAddress:            MainnetContractAddress,
		HyperionFullNodeIndexerURL: "https://api.hyperion.xyz",
		HyperionAPIHost:            "https://api.hyperion.xyz",
		ViewExecutor:               executor,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return sdk
}
