package server

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	maxContentLength = 1 << 10
)

type Request struct {
	Method      string
	URI         string
	HTTPVersion string
	Headers     []Header
	Body        string

	host          string
	contentLength uint64

	parseErr error
}

type Response struct {
	HTTPVersion  string
	StatusCode   int
	ReasonPhrase string
	Headers      []Header
	Body         string
}

type Header struct {
	Name  string
	Value string
}

func (srv *Server) handleConn(ctx context.Context, conn net.Conn) error {
	defer conn.Close() // nolint:errcheck

	stop := context.AfterFunc(ctx, func() {
		conn.SetReadDeadline(time.Now()) // nolint:errcheck
	})
	defer stop()

	req, err := readRequest(conn)
	if err != nil {
		return fmt.Errorf("failed to read request: %w", err)
	}

	res := srv.handleRequest(req)

	if err := srv.writeResponse(conn, res); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	log.Println(res.StatusCode, req.Method, req.URI)

	return conn.Close()
}

// readRequest parses a request from the stream.
// Errors when it fails to read for reasons other than malformed request.
func readRequest(r io.Reader) (*Request, error) {
	const (
		startLine uint8 = iota
		headerLine
		bodyLine
	)

	ltype := startLine
	req := new(Request)
	br := bufio.NewReader(r)

loop:
	for {
		line, isPrefix, err := br.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return req, nil
			}
			return nil, fmt.Errorf("failed to read line: %w", err)
		}
		if isPrefix {
			req.parseErr = errors.New("line too long")
			return req, nil
		}

		switch ltype {
		case startLine:
			parts := bytes.SplitN(line, []byte(" "), 3)
			if len(parts) != 3 {
				req.parseErr = errors.New("invalid start line format")
				return req, nil
			}
			req.Method = string(parts[0])
			req.URI = string(parts[1])
			req.HTTPVersion = string(parts[2])
			ltype = headerLine
		case headerLine:
			if len(line) == 0 {
				break loop // header end
			}
			parts := bytes.SplitN(line, []byte(":"), 2)
			if len(parts) != 2 {
				req.parseErr = errors.New("invalid header format")
				return req, nil
			}
			name := string(parts[0])
			val := string(bytes.TrimSpace(parts[1]))
			req.Headers = append(req.Headers, Header{name, val})
			switch strings.ToLower(name) {
			case "content-length":
				cl, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					req.parseErr = fmt.Errorf("failed to parse Content-Length: %w", err)
					return req, nil
				}
				req.contentLength = cl
			case "host":
				req.host = strings.TrimSuffix(val, "/")
			}
		}
	}

	if req.contentLength > maxContentLength {
		req.parseErr = errors.New("Content-Length too large")
		return req, nil
	}

	body := make([]byte, req.contentLength)
	_, err := io.ReadFull(br, body)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			req.parseErr = errors.New("body shorter than Content-Length")
			return req, nil
		}
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	req.Body = string(body)

	return req, nil
}

func (srv *Server) handleRequest(req *Request) *Response {
	res := new(Response)
	res.HTTPVersion = "HTTP/1.1"

	// always include these headers
	defer func() {
		res.Headers = append(res.Headers,
			Header{"Content-Type", "text/plain"},
			Header{"Location", fmt.Sprintf("http://%s%s", req.host, req.URI)},
			Header{"Content-Length", strconv.Itoa(len(res.Body))},
		)
	}()

	if req.parseErr != nil {
		res.StatusCode = 400
		res.ReasonPhrase = "Bad Request"
		return res
	}

	res.StatusCode = 200
	res.ReasonPhrase = "OK"

	switch req.URI {
	case "/":
		res.Body = "hi"
	case "/echo":
		switch req.Method {
		case "POST":
			res.Body = req.Body
		default:
			res.StatusCode = 405
			res.ReasonPhrase = "Method Not Allowed"
		}
	case "/store":
		switch req.Method {
		case "GET":
			res.Body = strings.Join(srv.db, "\n")
		case "POST":
			srv.db = append(srv.db, req.Body)
			res.Body = strings.Join(srv.db, "\n")
		default:
			res.StatusCode = 405
			res.ReasonPhrase = "Method Not Allowed"
		}
	default:
		res.StatusCode = 404
		res.ReasonPhrase = "Not Found"
	}

	return res
}

func (srv *Server) writeResponse(conn net.Conn, res *Response) error {
	w := bufio.NewWriter(conn)
	defer w.Flush() // nolint:errcheck

	// start line
	_, err := fmt.Fprintf(w, "%s %d %s\r\n", res.HTTPVersion, res.StatusCode, res.ReasonPhrase)
	if err != nil {
		return err
	}

	// headers
	for _, h := range res.Headers {
		_, err = fmt.Fprintf(w, "%s: %s\r\n", h.Name, h.Value)
		if err != nil {
			return err
		}
	}

	// body
	_, err = fmt.Fprintf(w, "\r\n%s", res.Body)
	if err != nil {
		return err
	}

	return nil
}
