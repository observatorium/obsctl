package main

import (
	"context"
	"os"
	"syscall"

	"github.com/observatorium/obsctl/pkg/cmd"
	"github.com/oklog/run"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := cmd.NewObsctlCmd(ctx)

	var g run.Group
	g.Add(func() error {
		return cmd.Execute()
	}, func(err error) {
		cancel()
	})

	// Listen for termination signals.
	g.Add(run.SignalHandler(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		os.Exit(1)

	}
}
