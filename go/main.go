package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	sport, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatalln("PORT not set")
	}

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		log.Fatalln("failed to parse port")
	}

	addr := fmt.Sprintf(":%d", port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("failed to dial tcp:", err)
	}
	defer listener.Close() // nolint:errcheck

	log.Println("server listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go handleConn(conn)
	}
}

type request struct {
	method      string
	uri         string
	httpVersion string
	headers     []header
}

type response struct {
	httpVersion  string
	statusCode   int
	reasonPhrase string
	headers      []header
	body         string
}

type header struct {
	name  string
	value string
}

func handleConn(conn net.Conn) {
	defer conn.Close() // nolint:errcheck
	s := bufio.NewScanner(conn)

	req := new(request)

	valid := true
	for s.Scan() {
		line := s.Text()
		if err := parseRequest(req, line); err != nil {
			valid = false
			break
		}
	}

	if err := s.Err(); err != nil {
		log.Fatalln("failed to read:", err)
	}

	res := handleRequest(req)
	if !valid {
		res.statusCode = 400
		res.reasonPhrase = "Bad Request"
	}

	// TODO: body writing
	w := bufio.NewWriter(conn)
	defer w.Flush() // nolint:errcheck
	_, err := fmt.Fprintf(w, "%s %d %s\r\n", res.httpVersion, res.statusCode, res.reasonPhrase)
	if err != nil {
		log.Fatalln("failed to write response", err)
	}
	for _, h := range res.headers {
		_, err = fmt.Fprintf(w, "%s: %s\r\n", h.name, h.value)
		if err != nil {
			log.Fatalln("failed to write response header", err)
		}
	}

	log.Println("request handled")
}

// TODO: body parsing
func parseRequest(req *request, line string) error {
	isStart := len(req.method) == 0
	if isStart {
		// start line
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			return errors.New("invalid start line format")
		}
		req.method = parts[0]
		req.uri = parts[1]
		req.httpVersion = parts[2]
	} else {
		// header line
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return errors.New("invalid header format")
		}
		req.headers = append(req.headers, header{parts[0], strings.TrimSpace(parts[1])})
	}

	return nil
}

func handleRequest(req *request) (res response) {
	res.httpVersion = "HTTP/1.1"

	// validate start line
	if req.method != "GET" {
		res.statusCode = 405
		res.reasonPhrase = "Method Not Allowed"
	} else {
		res.statusCode = 200
		res.reasonPhrase = "OK"
	}

	// validate uri
	switch req.uri {
	case "/":
		res.body = "hi"
	case "/users":
		res.body = "users"
	case "/posts":
		res.body = "posts"
	default:
		res.body = ""
		res.statusCode = 400
		res.reasonPhrase = "Bad Request"
	}

	// create location
	var loc strings.Builder
	loc.WriteString("http://")
	for _, h := range req.headers {
		if h.name == "Host" {
			loc.WriteString(strings.TrimSuffix(h.value, "/"))
			break
		}
	}
	loc.WriteString(req.uri)

	// req.headers
	res.headers = append(res.headers, header{"Content-Type", "plain/text"})
	res.headers = append(res.headers, header{"Location", loc.String()})

	return
}
