package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func newCompressRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(chimw.Compress(5))
	r.Get("/text", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(strings.Repeat("hello world ", 200)))
	})
	r.Get("/image", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 0x50, 0x4E, 0x47})
	})
	return r
}

func TestCompress_GzipEncoding(t *testing.T) {
	r := newCompressRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/text", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %q", ce)
	}

	gr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("response body is not valid gzip: %v", err)
	}
	defer gr.Close()

	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read decompressed body: %v", err)
	}

	if !strings.Contains(string(body), "hello world") {
		t.Errorf("expected body to contain 'hello world', got %q", string(body))
	}
}

func TestCompress_NoAcceptEncoding(t *testing.T) {
	r := newCompressRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/text", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if ce := w.Header().Get("Content-Encoding"); ce != "" {
		t.Errorf("expected no Content-Encoding, got %q", ce)
	}
}

func TestCompress_ImageNotCompressed(t *testing.T) {
	r := newCompressRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/image", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if ce := w.Header().Get("Content-Encoding"); ce != "" {
		t.Errorf("expected no Content-Encoding for image/png, got %q", ce)
	}
}
