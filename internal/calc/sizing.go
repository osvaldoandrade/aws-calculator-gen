package calc

func regionLabelFromCode(code string) string {
	switch code {
	case "us-east-1":
		return "US East (N. Virginia) [us-east-1]"
	case "us-east-2":
		return "US East (Ohio) [us-east-2]"
	case "us-west-2":
		return "US West (Oregon) [us-west-2]"
	case "eu-west-1":
		return "EU (Ireland) [eu-west-1]"
	case "sa-east-1":
		return "South America (S\u00e3o Paulo) [sa-east-1]"
	default:
		return code
	}
}
