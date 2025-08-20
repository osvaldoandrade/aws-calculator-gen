package pricing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculate(t *testing.T) {
	catFile, err := os.Open(filepath.Join("..", "..", "assets", "pricing-sample.yml"))
	if err != nil {
		t.Fatalf("catalog open: %v", err)
	}
	cat, err := LoadCatalog(catFile)
	if err != nil {
		t.Fatalf("catalog load: %v", err)
	}
	usageFile, err := os.Open("testdata/usage.yml")
	if err != nil {
		t.Fatalf("usage open: %v", err)
	}
	usage, err := LoadUsage(usageFile)
	if err != nil {
		t.Fatalf("usage load: %v", err)
	}
	breakdown, total := Calculate(cat, usage)
	if len(breakdown) == 0 || total <= 0 {
		t.Fatalf("unexpected result %v total %f", breakdown, total)
	}
	if int(total*1000) != 401816 {
		t.Fatalf("unexpected total %f", total)
	}
}
