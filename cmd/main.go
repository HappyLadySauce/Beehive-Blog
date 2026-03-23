package main

import (
	"context"
	"os"

	"k8s.io/component-base/cli"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app"
)

const (
	basename = "Beehive-Blog"
)

func main() {
	ctx := context.TODO()
	cmd := app.NewAPICommand(ctx, basename)
	code := cli.Run(cmd)
	os.Exit(code)
}

