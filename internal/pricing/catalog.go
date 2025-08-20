package pricing

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Catalog holds pricing by service and item.
type Catalog map[string]map[string]float64

// LoadCatalog reads catalog from YAML file.
func LoadCatalog(path string) (Catalog, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Catalog
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return c, nil
}
