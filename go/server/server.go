package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"golang.org/x/sync/errgroup"
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

func (srv *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return fmt.Errorf("failed to dial tcp: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		return listener.Close()
	})

	g.Go(func() error {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil // normal shutdown
				}
				return fmt.Errorf("failed to accept connection: %w", err)
			}

			g.Go(func() error {
				if err := srv.handleConn(ctx, conn); err != nil {
					log.Println("failed to handle connection: %w", err)
				}
				return nil
			})
		}
	})

	return g.Wait()
}
