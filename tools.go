//go:build tool

package headmail

import (
	_ "github.com/swaggo/swag/v2/cmd/swag"
	_ "github.com/swaggo/swag/v2/gen"
)

//go:generate swag init --v3.1 --parseDependency --parseInternal --dir ./cmd/server,./pkg --output ./docs
