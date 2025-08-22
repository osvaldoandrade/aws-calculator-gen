package awspc

import (
	"math"
	"testing"
)

func TestAssignUsage(t *testing.T) {
	lines := []usageLine{
		{price: 10},
		{price: 5},
		{price: 1},
	}
	assignUsage(lines, 26)
	total := 0.0
	for i := range lines {
		if lines[i].Amount == nil {
			t.Fatalf("expected amount for line %d", i)
		}
		total += *lines[i].Amount * lines[i].price
	}
	if math.Abs(total-26) > 1e-6 {
		t.Fatalf("expected total cost 26, got %f", total)
	}
	expected := 26.0 / 3
	for i := range lines {
		cost := *lines[i].Amount * lines[i].price
		if math.Abs(cost-expected) > 1e-6 {
			t.Fatalf("line %d cost %f, expected %f", i, cost, expected)
		}
	}
}
