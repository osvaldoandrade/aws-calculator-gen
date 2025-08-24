package command

import "testing"

func TestParseParams(t *testing.T) {
	got, err := ParseParams([]string{"a=1", "b=two"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["a"] != "1" || got["b"] != "two" {
		t.Fatalf("unexpected map: %#v", got)
	}
}

func TestParseParamsInvalid(t *testing.T) {
	_, err := ParseParams([]string{"bad"})
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}
