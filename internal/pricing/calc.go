package pricing

import "gopkg.in/yaml.v3"

// Usage model for pricing calculations.
type Usage struct {
	Region   string `yaml:"region"`
	Services struct {
		EC2 struct {
			Instances []struct {
				Family string  `yaml:"family"`
				Hours  float64 `yaml:"hours"`
			} `yaml:"instances"`
		} `yaml:"ec2"`
		EBS struct {
			GP3GBMonth        float64 `yaml:"gp3_gb_month"`
			IORequestsMillion float64 `yaml:"io_requests_millions"`
		} `yaml:"ebs"`
		S3 struct {
			StorageGBMonth     float64 `yaml:"storage_gb_month"`
			GetRequestsMillion float64 `yaml:"get_requests_millions"`
			PutRequestsMillion float64 `yaml:"put_requests_millions"`
			DataOutGB          float64 `yaml:"data_out_gb"`
		} `yaml:"s3"`
		RDS struct {
			Instance       string  `yaml:"instance"`
			Hours          float64 `yaml:"hours"`
			StorageGBMonth float64 `yaml:"storage_gb_month"`
		} `yaml:"rds"`
		Lambda struct {
			RequestsMillion float64 `yaml:"requests_millions"`
			GBSeconds       float64 `yaml:"gb_seconds"`
		} `yaml:"lambda"`
		CloudFront struct {
			DataOutGB float64 `yaml:"data_out_gb"`
		} `yaml:"cloudfront"`
	} `yaml:"services"`
}

// Calculator computes costs.
type Calculator struct {
	Catalog Catalog
}

// NewCalculator creates calculator with catalog.
func NewCalculator(c Catalog) *Calculator {
	return &Calculator{Catalog: c}
}

// Calculate total and breakdown.
func (c *Calculator) Calculate(u Usage) (float64, map[string]float64) {
	breakdown := make(map[string]float64)
	// EC2 instances
	for _, inst := range u.Services.EC2.Instances {
		rate := c.Catalog["ec2"][inst.Family]
		breakdown["ec2"] += rate * inst.Hours
	}
	// EBS
	breakdown["ebs"] += c.Catalog["ebs"]["gp3_gb_month"] * u.Services.EBS.GP3GBMonth
	breakdown["ebs"] += c.Catalog["ebs"]["io_requests_millions"] * u.Services.EBS.IORequestsMillion
	// S3
	breakdown["s3"] += c.Catalog["s3"]["storage_gb_month"] * u.Services.S3.StorageGBMonth
	breakdown["s3"] += c.Catalog["s3"]["get_requests_millions"] * u.Services.S3.GetRequestsMillion
	breakdown["s3"] += c.Catalog["s3"]["put_requests_millions"] * u.Services.S3.PutRequestsMillion
	breakdown["s3"] += c.Catalog["s3"]["data_out_gb"] * u.Services.S3.DataOutGB
	// RDS
	breakdown["rds"] += c.Catalog["rds"][u.Services.RDS.Instance] * u.Services.RDS.Hours
	breakdown["rds"] += c.Catalog["rds"]["storage_gb_month"] * u.Services.RDS.StorageGBMonth
	// Lambda
	breakdown["lambda"] += c.Catalog["lambda"]["requests_millions"] * u.Services.Lambda.RequestsMillion
	breakdown["lambda"] += c.Catalog["lambda"]["gb_seconds"] * u.Services.Lambda.GBSeconds
	// CloudFront
	breakdown["cloudfront"] += c.Catalog["cloudfront"]["data_out_gb"] * u.Services.CloudFront.DataOutGB
	var total float64
	for _, v := range breakdown {
		total += v
	}
	return total, breakdown
}

// ParseUsage loads Usage from YAML bytes.
func ParseUsage(b []byte) (Usage, error) {
	var u Usage
	if err := yaml.Unmarshal(b, &u); err != nil {
		return u, err
	}
	return u, nil
}
