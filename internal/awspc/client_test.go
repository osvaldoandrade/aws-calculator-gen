package awspc

import (
	"math"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	bcmtypes "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator/types"
)

func TestAssignUsage(t *testing.T) {
	lines := []usageLine{
		{BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{ServiceCode: aws.String("svc1")}, price: 10},
		{BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{ServiceCode: aws.String("svc2")}, price: 5},
		{BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{ServiceCode: aws.String("svc3")}, price: 1},
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

func TestAssignUsageSplitsWithinService(t *testing.T) {
	lines := []usageLine{
		{BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{ServiceCode: aws.String("svc1")}, price: 1},
		{BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{ServiceCode: aws.String("svc1")}, price: 1},
		{BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{ServiceCode: aws.String("svc2")}, price: 1},
	}
	assignUsage(lines, 6)
	if math.Abs(*lines[0].Amount*lines[0].price-1.5) > 1e-6 {
		t.Fatalf("line 0 cost %f, expected 1.5", *lines[0].Amount*lines[0].price)
	}
	if math.Abs(*lines[1].Amount*lines[1].price-1.5) > 1e-6 {
		t.Fatalf("line 1 cost %f, expected 1.5", *lines[1].Amount*lines[1].price)
	}
	if math.Abs(*lines[2].Amount*lines[2].price-3.0) > 1e-6 {
		t.Fatalf("line 2 cost %f, expected 3", *lines[2].Amount*lines[2].price)
	}
}
