package main

import (
	. "net/http"
	"net/http/httptest"
	"testing"
)

func testProbeHelper(status int) (*ProbeResult, error) {
	ts := httptest.NewServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		w.WriteHeader(status)
	}))
	defer ts.Close()

	r, err := probe(ts.URL)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func TestProbe(t *testing.T) {
	statuses := []struct {
		given int
		exp   string
	}{
		{200, "200 OK"},
		{300, ""},
		{400, ""},
		{500, ""},
	}

	for _, status := range statuses {
		r, err := testProbeHelper(status.given)
		if err != nil {
			if status.exp != "" {
				t.Fatal(err)
			} else {
				t.Logf("given %v exp %v r.status %v err %v", status.given, status.exp, nil, err)
				continue
			}
		} else if r.HttpStatus != status.exp {
			t.Fatalf("got %s: want: %s", r.HttpStatus, status.exp)
		}
		t.Logf("given %v exp %v r.status %v err %v", status.given, status.exp, r.HttpStatus, nil)
	}
}
