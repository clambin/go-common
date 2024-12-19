package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestTestServer(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		path            string
		wantStatusCode  int
		wantBody        string
		wantContentType string
	}{
		{"any method", http.MethodDelete, "/anymethod", http.StatusOK, "anymethod", "text/plain"},
		{"valid method", http.MethodGet, "/getorpost", http.StatusOK, "getorpost", "text/plain"},
		{"invalid method", http.MethodDelete, "/getorpost", http.StatusMethodNotAllowed, "405 Method Not Allowed\n", "text/plain; charset=utf-8"},
		{"invalid path", http.MethodGet, "/invalid", http.StatusNotFound, "404 Not Found\n", "text/plain; charset=utf-8"},
		{"json", http.MethodGet, "/json", http.StatusOK, "{\"Value\":1}\n", "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := map[string]Path{
				"/anymethod": {Methods: nil, Body: "anymethod"},
				"/getonly":   {Methods: []string{http.MethodGet}, Body: []byte("getonly")},
				"/getorpost": {Methods: []string{http.MethodGet, http.MethodPost}, Body: "getorpost"},
				"/json":      {Methods: nil, Body: struct{ Value int }{Value: 1}},
			}
			ts := TestServer{Paths: paths}

			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			ts.ServeHTTP(w, req)
			if tt.wantStatusCode != w.Code {
				t.Errorf("want status code %v, got %v", tt.wantStatusCode, w.Code)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("want body %q, got %q", tt.wantBody, w.Body.String())
			}
			if w.Header().Get("Content-Type") != tt.wantContentType {
				t.Errorf("want content type %q, got %q", tt.wantContentType, w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	for _, obj := range []any{
		"foo",
		[]byte("bar"),
		struct{ Value int }{Value: 242},
	} {
		_ = json.NewEncoder(os.Stdout).Encode(obj)
	}
}
