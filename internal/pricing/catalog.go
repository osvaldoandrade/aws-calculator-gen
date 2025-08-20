package pricing

import (
	"io"

	"gopkg.in/yaml.v3"
)

// Catalog holds price per unit for services.
type Catalog struct {
	EC2        map[string]float64 `yaml:"ec2"`
	EBS        map[string]float64 `yaml:"ebs"`
	S3         map[string]float64 `yaml:"s3"`
	RDS        map[string]float64 `yaml:"rds"`
	Lambda     map[string]float64 `yaml:"lambda"`
	CloudFront map[string]float64 `yaml:"cloudfront"`
}

// LoadCatalog decodes pricing from reader.
func LoadCatalog(r io.Reader) (*Catalog, error) {
	var c Catalog
	if err := yaml.NewDecoder(r).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
