package hyperion

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type payloadSnapshot struct {
	Function          string   `json:"function"`
	TypeArguments     []string `json:"typeArguments"`
	FunctionArguments []any    `json:"functionArguments"`
}

func normalizePayloadSnapshot(payload EntryFunctionPayload) payloadSnapshot {
	args := make([]any, 0, len(payload.FunctionArguments))
	for _, arg := range payload.FunctionArguments {
		args = append(args, normalizeGoldenValue(arg))
	}
	return payloadSnapshot{
		Function:          payload.Function,
		TypeArguments:     payload.TypeArguments,
		FunctionArguments: args,
	}
}

func assertGoldenJSON(t *testing.T, path string, actual any) {
	t.Helper()

	expectedBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden fixture %s: %v", path, err)
	}

	expected, err := canonicalJSON(expectedBytes)
	if err != nil {
		t.Fatalf("decode golden fixture %s: %v", path, err)
	}
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("marshal actual value: %v", err)
	}
	got, err := canonicalJSON(actualBytes)
	if err != nil {
		t.Fatalf("decode actual value: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		gotPretty, _ := json.MarshalIndent(got, "", "  ")
		expectedPretty, _ := json.MarshalIndent(expected, "", "  ")
		t.Fatalf("golden fixture mismatch for %s\nexpected:\n%s\nactual:\n%s", path, expectedPretty, gotPretty)
	}
}

func writeFixture(w http.ResponseWriter, path string) error {
	fixture, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if _, err := w.Write(fixture); err != nil {
		return err
	}
	return nil
}

func canonicalJSON(input []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()
	var out any
	if err := decoder.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizeGoldenValue(value any) any {
	switch typed := value.(type) {
	case int64:
		if typed > 100*365*24*60*60 {
			return "<deadline>"
		}
		return typed
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, normalizeGoldenValue(item))
		}
		return out
	case []string:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	default:
		return value
	}
}
