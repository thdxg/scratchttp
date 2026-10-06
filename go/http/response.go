package http

import (
	"bufio"
	"fmt"
	"io"
)

type Response struct {
	HTTPVersion  string
	StatusCode   int
	ReasonPhrase string
	Headers      []Header
	Body         string
}

func WriteResponse(w io.Writer, res *Response) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush() // nolint:errcheck

	// start line
	_, err := fmt.Fprintf(bw, "%s %d %s\r\n", res.HTTPVersion, res.StatusCode, res.ReasonPhrase)
	if err != nil {
		return err
	}

	// headers
	for _, h := range res.Headers {
		_, err = fmt.Fprintf(bw, "%s: %s\r\n", h.Name, h.Value)
		if err != nil {
			return err
		}
	}

	// body
	_, err = fmt.Fprintf(bw, "\r\n%s", res.Body)
	if err != nil {
		return err
	}

	return nil
}
