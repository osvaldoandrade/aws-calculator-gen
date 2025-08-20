package pricing

import (
	"io"

	"gopkg.in/yaml.v3"
)

// Usage represents workload usage metrics.
type Usage struct {
	Region   string                 `yaml:"region"`
	Services map[string]interface{} `yaml:"services"`
}

// LoadUsage decodes usage from reader.
func LoadUsage(r io.Reader) (*Usage, error) {
	var u Usage
	if err := yaml.NewDecoder(r).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}
