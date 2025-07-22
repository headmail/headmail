//go:build tool

package headmail

import (
	_ "github.com/swaggo/swag/cmd/swag"
)

//go:generate swag init --dir ./cmd/server,./pkg --output ./docs
