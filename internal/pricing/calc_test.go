package pricing

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculate(t *testing.T) {
	catalog, err := LoadCatalog(filepath.Join("..", "..", "assets", "pricing-sample.yml"))
	require.NoError(t, err)
	calc := NewCalculator(catalog)
	b, err := ioutil.ReadFile(filepath.Join("testdata", "usage.yml"))
	require.NoError(t, err)
	u, err := ParseUsage(b)
	require.NoError(t, err)
	total, breakdown := calc.Calculate(u)
	require.Greater(t, total, 0.0)
	require.Len(t, breakdown, 6)
}
