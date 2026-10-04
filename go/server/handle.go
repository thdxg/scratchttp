package server

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type Request struct {
	Method      string
	URI         string
	HTTPVersion string
	Headers     []Header
	Body        string

	// ltype is one of 0 for start line, 1 for header line, and 2 for body
	// used for tracking parse status
	ltype uint8
}

const (
	startLine uint8 = iota
	headerLine
	bodyLine
)

type Response struct {
	HttpVersion  string
	StatusCode   int
	ReasonPhrase string
	Headers      []Header
	Body         string
}

type Header struct {
	Name  string
	Value string
}

func (srv *Server) handleConn(conn net.Conn) {
	defer conn.Close() // nolint:errcheck
	s := bufio.NewScanner(conn)

	req := new(Request)

	valid := true
	for s.Scan() {
		line := s.Text()
		if err := srv.parseRequest(req, line); err != nil {
			valid = false
			break
		}
	}

	if err := s.Err(); err != nil {
		log.Fatalln("failed to read from conn:", err)
	}

	res := srv.handleRequest(req)
	if !valid {
		res.StatusCode = 400
		res.ReasonPhrase = "Bad Request"
	}

	if err := srv.writeResponse(conn, res); err != nil {
		log.Fatalln("failed to write response", err)
	}

	log.Println(res.StatusCode, req.Method, req.URI)
}

func (srv *Server) parseRequest(req *Request, line string) error {
	if req.ltype == startLine && len(req.Method) != 0 {
		req.ltype = headerLine
	} else if line == "" {
		req.ltype = bodyLine
		return nil
	}

	switch req.ltype {
	case startLine:
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			return errors.New("invalid start line format")
		}
		req.Method = parts[0]
		req.URI = parts[1]
		req.HTTPVersion = parts[2]
	case headerLine:
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return errors.New("invalid header format")
		}
		req.Headers = append(req.Headers, Header{parts[0], strings.TrimSpace(parts[1])})
	case bodyLine:
		req.Body = line
	}

	return nil
}

func (srv *Server) handleRequest(req *Request) *Response {
	res := new(Response)

	res.HttpVersion = "HTTP/1.1"

	// validate start line
	res.StatusCode = 200
	res.ReasonPhrase = "OK"

	// validate uri
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
		res.Body = ""
		res.StatusCode = 400
		res.ReasonPhrase = "Bad Request"
	}

	// create location
	var loc strings.Builder
	loc.WriteString("http://")
	for _, h := range req.Headers {
		if h.Name == "Host" {
			loc.WriteString(strings.TrimSuffix(h.Value, "/"))
			break
		}
	}
	loc.WriteString(req.URI)

	// req.headers
	res.Headers = append(res.Headers, Header{"Content-Type", "plain/text"})
	res.Headers = append(res.Headers, Header{"Location", loc.String()})

	return res
}

func (srv *Server) writeResponse(conn net.Conn, res *Response) error {
	w := bufio.NewWriter(conn)
	defer w.Flush() // nolint:errcheck

	// start line
	_, err := fmt.Fprintf(w, "%s %d %s\r\n", res.HttpVersion, res.StatusCode, res.ReasonPhrase)
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
