package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"scratchttp/http"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	addr   string
	db     *DB
	routes map[string]Handler
}

func New(addr string) *Server {
	srv := new(Server)

	srv.addr = addr
	srv.db = new(DB)
	srv.routes = map[string]Handler{
		"/":      srv.HandleIndex,
		"/echo":  srv.HandleEcho,
		"/store": srv.HandleStore,
	}

	return srv
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
					log.Println("failed to handle connection:", err)
				}
				return nil
			})
		}
	})

	return g.Wait()
}

func (srv *Server) handleConn(ctx context.Context, conn net.Conn) error {
	defer conn.Close() // nolint:errcheck

	stop := context.AfterFunc(ctx, func() {
		conn.SetReadDeadline(time.Now()) // nolint:errcheck
	})
	defer stop()

	req, err := http.ReadRequest(conn)
	if err != nil {
		return fmt.Errorf("failed to read request: %w", err)
	}

	res := srv.handleRequest(req)

	if err := http.WriteResponse(conn, res); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	log.Println(res.StatusCode, req.Method, req.URI)

	return conn.Close()
}

func (srv *Server) handleRequest(req *http.Request) *http.Response {
	res := new(http.Response)

	defer func() {
		res.HTTPVersion = "HTTP/1.1"
		res.Headers = append(res.Headers,
			http.Header{Name: "Content-Type", Value: "text/plain"},
			http.Header{Name: "Location", Value: fmt.Sprintf("http://%s%s", req.Host, req.URI)},
			http.Header{Name: "Content-Length", Value: strconv.Itoa(len(res.Body))},
		)
	}()

	if req.ParseErr != nil {
		res.StatusCode = 400
		res.ReasonPhrase = "Bad Request"
		return res
	}

	handler, ok := srv.routes[req.URI]
	if !ok {
		res.StatusCode = 404
		res.ReasonPhrase = "Not Found"
		return res
	}

	res = handler(req)

	return res
}
