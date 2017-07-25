package wiki_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suzuken/wiki"
)

func TestGETHandler(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
	ts := httptest.NewServer(wiki.GET(h))
	if _, err := http.Get(ts.URL); err != nil {
		t.Fatalf("GET failed: %s", err)
	}
	resp, err := http.Post(ts.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("POST failed: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("want %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestPOSTHandler(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
	ts := httptest.NewServer(wiki.POST(h))
	if _, err := http.Post(ts.URL, "application/json", nil); err != nil {
		t.Fatalf("POST failed: %s", err)
	}
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET failed: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("want %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}
