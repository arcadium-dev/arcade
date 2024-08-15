//go:build tools

package arcade

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
