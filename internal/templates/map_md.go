package templates

import _ "embed"

//go:embed MAP.md
var defaultMAP string

func MAPTemplate() string { return defaultMAP }
