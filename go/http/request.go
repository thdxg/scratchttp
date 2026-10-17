package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	maxRequestContentLength uint64 = 1 << 10
)

type Method string

const (
	MethodGet    Method = "GET"
	MethodPost   Method = "POST"
	MethodPut    Method = "PUT"
	MethodDelete Method = "DELETE"
)

type Request struct {
	Method      Method
	URI         string
	HTTPVersion string
	Headers     []Header
	Body        string

	Host          string
	ContentLength uint64

	ParseErr error
}

// ReadRequest parses a request from the stream.
// Errors when it fails to read for reasons other than malformed request.
func ReadRequest(r io.Reader) (*Request, error) {
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
			req.ParseErr = errors.New("line too long")
			return req, nil
		}

		switch ltype {
		case startLine:
			parts := bytes.SplitN(line, []byte(" "), 3)
			if len(parts) != 3 {
				req.ParseErr = errors.New("malformed start line")
				return req, nil
			}
			req.URI = string(parts[1])
			req.HTTPVersion = string(parts[2])
			switch Method(parts[0]) {
			case MethodGet, MethodPost, MethodPut, MethodDelete:
				req.Method = Method(parts[0])
			default:
				req.ParseErr = errors.New("malformed start line")
				return req, nil
			}
			ltype = headerLine
		case headerLine:
			if len(line) == 0 {
				ltype = bodyLine // nolint:ineffassign
				break loop       // header end
			}
			parts := bytes.SplitN(line, []byte(":"), 2)
			if len(parts) != 2 {
				req.ParseErr = errors.New("malformed header")
				return req, nil
			}
			name := string(parts[0])
			val := string(bytes.TrimSpace(parts[1]))
			req.Headers = append(req.Headers, Header{name, val})
			switch strings.ToLower(name) {
			case "content-length":
				cl, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					req.ParseErr = fmt.Errorf("failed to parse Content-Length: %w", err)
					return req, nil
				}
				req.ContentLength = cl
			case "host":
				req.Host = strings.TrimSuffix(val, "/")
			}
		}
	}

	if req.ContentLength > maxRequestContentLength {
		req.ParseErr = errors.New("Content-Length too large")
		return req, nil
	}

	body := make([]byte, req.ContentLength)
	_, err := io.ReadFull(br, body)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			req.ParseErr = errors.New("body shorter than Content-Length")
			return req, nil
		}
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	req.Body = string(body)

	return req, nil
}
