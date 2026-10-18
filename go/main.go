package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"scratchttp/server"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr, ok := os.LookupEnv("ADDRESS")
	if !ok {
		fmt.Fprintln(os.Stderr, "ADDRESS not set")
		os.Exit(1)
	}

	srv := server.New(addr)

	if err := srv.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "failed to run server:", err)
		os.Exit(1)
	}
}
