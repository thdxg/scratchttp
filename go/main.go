package main

import (
	"fmt"
	"log"
	"os"
	"scratchttp/server"
	"strconv"
)

func main() {
	sport, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatalln("PORT not set")
	}

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		log.Fatalln("failed to parse port", err)
	}

	addr := fmt.Sprintf(":%d", port)

	srv := server.New(addr)

	if err := srv.Run(); err != nil {
		log.Fatalln("failed to run server", err)
	}
}
