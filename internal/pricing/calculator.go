package pricing

// Calculate computes cost per service and total.
func Calculate(cat *Catalog, u *Usage) (map[string]float64, float64) {
	b := make(map[string]float64)

	if s, ok := u.Services["ec2"].(map[string]interface{}); ok {
		if insts, ok := s["instances"].([]interface{}); ok {
			for _, i := range insts {
				m := i.(map[string]interface{})
				family := m["family"].(string)
				hours := toF64(m["hours"])
				price := cat.EC2[family]
				b["ec2"] += hours * price
			}
		}
	}

	if s, ok := u.Services["ebs"].(map[string]interface{}); ok {
		b["ebs"] += toF64(s["gp3_gb_month"]) * cat.EBS["gp3_gb_month"]
		b["ebs"] += toF64(s["io_requests_millions"]) * cat.EBS["io_requests_million"]
	}

	if s, ok := u.Services["s3"].(map[string]interface{}); ok {
		b["s3"] += toF64(s["storage_gb_month"]) * cat.S3["storage_gb_month"]
		b["s3"] += toF64(s["get_requests_millions"]) * cat.S3["get_requests_million"]
		b["s3"] += toF64(s["put_requests_millions"]) * cat.S3["put_requests_million"]
		b["s3"] += toF64(s["data_out_gb"]) * cat.S3["data_out_gb"]
	}

	if s, ok := u.Services["rds"].(map[string]interface{}); ok {
		inst := s["instance"].(string)
		b["rds"] += toF64(s["hours"]) * cat.RDS[inst]
		b["rds"] += toF64(s["storage_gb_month"]) * cat.RDS["storage_gb_month"]
	}

	if s, ok := u.Services["lambda"].(map[string]interface{}); ok {
		b["lambda"] += toF64(s["requests_millions"]) * cat.Lambda["requests_million"]
		b["lambda"] += toF64(s["gb_seconds"]) * cat.Lambda["gb_seconds"]
	}

	if s, ok := u.Services["cloudfront"].(map[string]interface{}); ok {
		b["cloudfront"] += toF64(s["data_out_gb"]) * cat.CloudFront["data_out_gb"]
	}

	total := 0.0
	for _, v := range b {
		total += v
	}
	return b, total
}

func toF64(v interface{}) float64 {
	switch t := v.(type) {
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case float64:
		return t
	default:
		return 0
	}
}
