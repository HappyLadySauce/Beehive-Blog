package main

import (
	"context"
	"os"

	"k8s.io/component-base/cli"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app"
)

func main() {
	ctx := context.TODO()
	cmd := app.NewAPICommand(ctx)
	code := cli.Run(cmd)
	os.Exit(code)
}

