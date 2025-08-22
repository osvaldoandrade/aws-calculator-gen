package awspc

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	bcmtypes "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator/types"
	"math"
	"testing"
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

func TestAssignUsageLakeProfile(t *testing.T) {
	lines := defaultEntries("us-east-1", "lake")
	assignUsage(lines, 300000)
	services := map[string]float64{}
	total := 0.0
	for _, l := range lines {
		if l.Amount == nil || l.price <= 0 {
			continue
		}
		cost := *l.Amount * l.price
		svc := aws.ToString(l.ServiceCode)
		services[svc] += cost
		total += cost
	}
	if len(services) != 6 {
		t.Fatalf("expected 6 services, got %d", len(services))
	}
	perService := 300000.0 / 6
	for svc, cost := range services {
		if math.Abs(cost-perService) > 1 {
			t.Fatalf("service %s cost %f, expected %f", svc, cost, perService)
		}
	}
	if math.Abs(total-300000) > 1 {
		t.Fatalf("total cost %f, expected 300000", total)
	}
}

func TestAssignUsageTransactionalProfile(t *testing.T) {
	lines := defaultEntries("us-east-1", "transactional")
	assignUsage(lines, 120000)
	services := map[string]float64{}
	total := 0.0
	for _, l := range lines {
		if l.Amount == nil || l.price <= 0 {
			continue
		}
		cost := *l.Amount * l.price
		svc := aws.ToString(l.ServiceCode)
		services[svc] += cost
		total += cost
	}
	if len(services) != 6 {
		t.Fatalf("expected 6 services, got %d", len(services))
	}
	perService := 120000.0 / 6
	for svc, cost := range services {
		if math.Abs(cost-perService) > 1 {
			t.Fatalf("service %s cost %f, expected %f", svc, cost, perService)
		}
	}
	if math.Abs(total-120000) > 1 {
		t.Fatalf("total cost %f, expected 120000", total)
	}

	prefix := usagePrefix("us-east-1")
	for _, l := range lines {
		switch aws.ToString(l.ServiceCode) {
		case "AWSEvents":
			if aws.ToString(l.UsageType) != prefix+"Event-64K-Chunks" {
				t.Fatalf("unexpected events usage type %s", aws.ToString(l.UsageType))
			}
		case "AmazonStates":
			if aws.ToString(l.UsageType) != prefix+"StateTransition" || aws.ToString(l.Operation) != "" {
				t.Fatalf("unexpected states entry: %+v", l)
			}
		case "AmazonRDS":
			if aws.ToString(l.Operation) != "CreateDBInstance:0014" {
				t.Fatalf("unexpected rds operation %s", aws.ToString(l.Operation))
			}
		case "AmazonElastiCache":
			if aws.ToString(l.Operation) != "CreateCacheCluster:0002" {
				t.Fatalf("unexpected elasticache operation %s", aws.ToString(l.Operation))
			}
		}
	}
}
