package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTestServer(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		wantBody       string
	}{
		{"any method", http.MethodDelete, "/anymethod", http.StatusOK, "anymethod"},
		{"valid method", http.MethodGet, "/getorpost", http.StatusOK, "getorpost"},
		{"invalid method", http.MethodDelete, "/getorpost", http.StatusMethodNotAllowed, "405 Method Not Allowed\n"},
		{"invalid path", http.MethodGet, "/invalid", http.StatusNotFound, "404 Not Found\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := map[string]Path{
				"/anymethod": {Methods: nil, Body: []byte("anymethod")},
				"/getonly":   {Methods: []string{http.MethodGet}, Body: []byte("getonly")},
				"/getorpost": {Methods: []string{http.MethodGet, http.MethodPost}, Body: []byte("getorpost")},
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
		})
	}
}
