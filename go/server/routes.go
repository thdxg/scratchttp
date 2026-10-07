package server

import (
	"encoding/json"
	"scratchttp/http"
)

type Handler func(req *http.Request) (res *http.Response)

func (srv *Server) HandleIndex(req *http.Request) *http.Response {
	res := new(http.Response)

	res.Headers = append(res.Headers,
		http.Header{Name: "Content-Type", Value: "text/plain"},
	)

	switch req.Method {
	case "GET":
		res.StatusCode = 200
		res.ReasonPhrase = "OK"
		res.Body = "hi"
	default:
		res.StatusCode = 405
		res.ReasonPhrase = "Method Not Allowed"
	}

	return res
}

func (srv *Server) HandleEcho(req *http.Request) *http.Response {
	res := new(http.Response)

	res.Headers = append(res.Headers,
		http.Header{Name: "Content-Type", Value: "text/plain"},
	)

	switch req.Method {
	case "POST":
		res.Body = req.Body
		res.StatusCode = 200
		res.ReasonPhrase = "OK"
	default:
		res.StatusCode = 405
		res.ReasonPhrase = "Method Not Allowed"
	}

	return res
}

func (srv *Server) HandleStore(req *http.Request) *http.Response {
	res := new(http.Response)

	res.Headers = append(res.Headers,
		http.Header{Name: "Content-Type", Value: "application/json"},
	)

	switch req.Method {
	case "GET":
		res.StatusCode = 200
		res.ReasonPhrase = "OK"
		data, err := json.Marshal(srv.db.Get())
		if err != nil {
			res.StatusCode = 500
			res.ReasonPhrase = "Internal Server Error"
		}
		res.Body = string(data)
	case "POST":
		srv.db.Add(req.Body)
		res.StatusCode = 200
		res.ReasonPhrase = "OK"
		data, err := json.Marshal(srv.db.Get())
		if err != nil {
			res.StatusCode = 500
			res.ReasonPhrase = "Internal Server Error"
		}
		res.Body = string(data)
	default:
		res.StatusCode = 405
		res.ReasonPhrase = "Method Not Allowed"
	}

	return res
}
