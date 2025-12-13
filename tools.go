//go:build tools
// +build tools

// https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md
package main

import (
	_ "github.com/swaggo/swag/cmd/swag"
	_ "github.com/vektra/mockery/v3"
)
