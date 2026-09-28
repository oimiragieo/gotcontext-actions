package main

import (
	_ "embed"

	"github.com/oimiragieo/gotcontext-actions/internal/cli"
	"github.com/oimiragieo/gotcontext-actions/internal/common"
)

//go:embed VERSION
var version string

func main() {
	ctx, cancel := common.CreateGracefulJobCancellationContext()
	defer cancel()
	cli.Execute(ctx, version)
}
