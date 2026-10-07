package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func Test_main(t *testing.T) {
	path, ok := os.LookupEnv("TESTDATA")
	if !ok {
		t.Fatal("TESTDATA not set")
	}

	addr, ok := os.LookupEnv("ADDRESS")
	if !ok {
		t.Fatal("ADDRESS not set")
	}

	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	data := make([]string, 0)
	if err := json.Unmarshal(f, &data); err != nil {
		t.Fatalf("failed to unmarshal testdata: %v", err)
	}

	go main()
	time.Sleep(time.Second) // wait for server

	c := http.DefaultClient
	url := fmt.Sprintf("http://%s%s", addr, "/store")

	t.Run("post", func(t *testing.T) {
		for i, d := range data {
			t.Run(fmt.Sprintf("request %d", i), func(t *testing.T) {
				t.Parallel()
				res, err := c.Post(url, "text/plain", strings.NewReader(d))
				if err != nil {
					t.Fatalf("request failed: %v", err)
				}
				defer res.Body.Close() // nolint:errcheck

				if res.StatusCode != 200 {
					t.Fatalf("unexpected status: %d", res.StatusCode)
				}
			})
		}
	})

	res, err := c.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close() // nolint:errcheck

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	expected := slices.Clone(data)
	slices.Sort(expected)

	actual := make([]string, 0)
	if err := json.Unmarshal(body, &actual); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if !slices.Equal(expected, actual) {
		t.Fatalf("\nexpected: %s\n  actual: %s",
			strings.Join(expected, ","), strings.Join(actual, ","))
	}
}
