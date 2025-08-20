package awspc

import "testing"

func TestAssignUsage(t *testing.T) {
	lines := []usageLine{
		{price: 10},
		{price: 5},
		{price: 1},
	}
	assignUsage(lines, 26)
	if lines[0].Amount == nil || *lines[0].Amount != 2 {
		t.Fatalf("expected first line 2 units, got %v", lines[0].Amount)
	}
	if lines[1].Amount == nil || *lines[1].Amount != 1 {
		t.Fatalf("expected second line 1 unit, got %v", lines[1].Amount)
	}
	if lines[2].Amount == nil || *lines[2].Amount != 1 {
		t.Fatalf("expected third line 1 unit, got %v", lines[2].Amount)
	}
}
