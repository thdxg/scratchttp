package main

import (
	"context"
	"log"
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
		log.Fatalln("ADDRESS not set")
	}

	srv := server.New(addr)

	if err := srv.Run(ctx); err != nil {
		log.Fatalln("failed to run server:", err)
	}
}
