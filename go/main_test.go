package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

// test checks either the exact response (Res) or, when the body depends on
// what other concurrent requests did, only the status line (Status).
type test struct {
	Req    string `json:"req"`
	Res    string `json:"res"`
	Status string `json:"status"`
}

func Test_main(t *testing.T) {
	path, ok := os.LookupEnv("TESTDATA")
	if !ok {
		t.Fatal("TESTDATA not set")
	}

	sport, ok := os.LookupEnv("PORT")
	if !ok {
		t.Fatal("PORT not set")
	}

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	// Phases run in order; the tests inside a phase run in parallel.
	var phases [][]test
	if err := json.Unmarshal(data, &phases); err != nil {
		t.Fatalf("failed to unmarshal testdata: %v", err)
	}

	addr := fmt.Sprintf("localhost:%d", port)

	go main()
	waitForServer(t, addr)

	for p, phase := range phases {
		t.Run(fmt.Sprintf("phase%d", p), func(t *testing.T) {
			for i, tt := range phase {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					t.Parallel()

					res := roundTrip(t, addr, tt.Req)

					if tt.Status != "" {
						status, _, _ := bytes.Cut(res, []byte("\r\n"))
						if string(status) != tt.Status {
							t.Errorf("request: %q\nwant status: %q\n got status: %q", tt.Req, tt.Status, status)
						}
						return
					}

					if string(res) != tt.Res {
						t.Errorf("request: %q\nwant: %q\n got: %q", tt.Req, tt.Res, res)
					}
				})
			}
		})
	}
}

func roundTrip(t *testing.T, addr, req string) []byte {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to dial tcp: %v", err)
	}
	defer conn.Close() // nolint:errcheck

	conn.SetDeadline(time.Now().Add(2 * time.Second)) // nolint:errcheck

	if _, err := io.WriteString(conn, req); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if err := conn.(*net.TCPConn).CloseWrite(); err != nil {
		t.Fatalf("failed to close write side: %v", err)
	}

	res, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	return res
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close() // nolint:errcheck
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server not listening on %s", addr)
}
