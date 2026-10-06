package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

type test struct {
	Req string `json:"req"`
	Res string `json:"res"`
}

func Test_main(t *testing.T) {
	path, ok := os.LookupEnv("TESTDATA")
	if !ok {
		t.Fatalf("TESTDATA not set")
	}

	sport, ok := os.LookupEnv("PORT")
	if !ok {
		t.Fatal("PORT not set")
	}

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open testdata: %v", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	tests := make([]test, 0)

	if err := json.Unmarshal(data, &tests); err != nil {
		t.Fatalf("failed to unmarshal testdata: %v", err)
	}

	addr := fmt.Sprintf(":%d", port)

	go main()
	waitForServer(t, addr)

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				t.Fatalf("failed to dial tcp: %v", err)
			}
			defer conn.Close() // nolint:errcheck

			_, err = io.WriteString(conn, tt.Req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}

			if err := conn.(*net.TCPConn).CloseWrite(); err != nil {
				t.Fatalf("failed to close write side: %v", err)
			}

			res, err := io.ReadAll(conn)
			if err != nil {
				t.Fatalf("failed to read response: %v", err)
			}

			if string(res) != tt.Res {
				t.Logf("wanted: %s\ngot: %s\n", tt.Res, res)
				t.Fail()
			}
		})
	}

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
