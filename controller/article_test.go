package controller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suzuken/wiki/controller"
)

type testHandler func(w http.ResponseWriter, r *http.Request) error

func (h testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// NOTE ignore error
	h(w, r)
}

func TestNotfound(t *testing.T) {
	article := &controller.Article{}
	rec := httptest.NewRecorder()
	handler := testHandler(article.Root)

	req, err := http.NewRequest("GET", "/hoge", nil)
	if err != nil {
		t.Fatalf("make request failed: %s", err)
	}
	handler.ServeHTTP(rec, req)

	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("want %d, got %d", http.StatusNotFound, status)
	}
}

func TestCSRFProtection(t *testing.T) {
	// POST to article
	// maybe block
}
