package server

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	addr string
	db   []string
}

func New(addr string) *Server {
	return &Server{
		addr: addr,
		db:   make([]string, 0),
	}
}

func (srv *Server) Run() error {
	listener, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return fmt.Errorf("failed to dial tcp: %w", err)
	}
	defer listener.Close() // nolint:errcheck

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go srv.handleConn(conn)
	}
}
