package incentives

import "testing"

func TestComputeMAPFunding(t *testing.T) {
	plan := ComputeMAPFunding(250000)
	if plan.Tier != "300k" {
		t.Fatalf("expected tier 300k got %s", plan.Tier)
	}
	if plan.CapAmount != 250000*0.10 {
		t.Fatalf("unexpected cap amount %f", plan.CapAmount)
	}
}

func TestComputeMAPFundingAbove(t *testing.T) {
	plan := ComputeMAPFunding(700000)
	if plan.Tier != ">600k" {
		t.Fatalf("expected >600k got %s", plan.Tier)
	}
}
