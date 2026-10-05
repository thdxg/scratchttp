package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"scratchttp/server"
	"strconv"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	sport, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatalln("PORT not set")
	}

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		log.Fatalln("failed to parse port:", err)
	}

	addr := fmt.Sprintf(":%d", port)
	srv := server.New(addr)

	if err := srv.Run(ctx); err != nil {
		log.Fatalln("failed to run server:", err)
	}
}
