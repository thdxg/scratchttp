package server

import (
	"scratchttp/http"
	"strings"
)

type Handler func(req *http.Request) (res *http.Response)

func (srv *Server) HandleIndex(req *http.Request) *http.Response {
	res := new(http.Response)

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

	switch req.Method {
	case "GET":
		res.StatusCode = 200
		res.ReasonPhrase = "OK"
		res.Body = strings.Join(srv.db.Get(), "\n")
	case "POST":
		srv.db.Add(req.Body)
		res.StatusCode = 200
		res.ReasonPhrase = "OK"
		res.Body = strings.Join(srv.db.Get(), "\n")
	default:
		res.StatusCode = 405
		res.ReasonPhrase = "Method Not Allowed"
	}

	return res
}
