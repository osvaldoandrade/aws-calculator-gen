package command

import (
	"bytes"
	"context"
	"testing"

	"github.com/example/seidor-tools/internal/calc"
)

func TestMapCommandName(t *testing.T) {
	if NewMapCommand().Name() != "map" {
		t.Fatalf("unexpected name")
	}
}

func TestMapCommandRun(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := &MapCommand{
		out:          buf,
		startSpinner: nil,
		runOrchestrator: func(ctx context.Context, o calc.Orchestrator) (calc.Result, error) {
			return calc.Result{
				ShareURL:      "https://example.com",
				RegionLabel:   "us-east-1",
				InstanceType:  "t3.micro",
				Count:         1,
				AchievedMRR:   100,
				RelativeError: 0.01,
			}, nil
		},
	}

	params := map[string]string{
		"customer":    "ACME",
		"description": "Test",
		"region":      "us-east-1",
		"arr":         "1200",
	}

	err := cmd.Run(context.Background(), params)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("https://example.com")) {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestGetStringParam(t *testing.T) {
	v, err := getStringParam(map[string]string{"key": "value"}, "key", "prompt")
	if err != nil || v != "value" {
		t.Fatalf("expected value, got %s err %v", v, err)
	}
}

func TestGetFloatParam(t *testing.T) {
	v, err := getFloatParam(map[string]string{"num": "1.5"}, "num", "prompt")
	if err != nil || v != 1.5 {
		t.Fatalf("expected 1.5, got %f err %v", v, err)
	}
}
