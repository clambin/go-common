package testutils

import (
	"cmp"
	"encoding/json"
	"maps"
	"net/http"
	"sync"
)

var _ http.Handler = &TestServer{}

// TestServer is a dummy http.Handler that can be used to create a httptest.Server.
// TestServer serves HTTP requests, using [Path] to determine how to reply, by looking up r.URL.Path in Paths.
// If no Path matches the request, TestServer returns HTTP 404
type TestServer struct {
	Paths map[string]Path
	lock  sync.RWMutex
	calls map[string]int
}

func (t *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, ok := t.Paths[r.URL.Path]
	if !ok {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	if !path.validMethod(r) {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	t.count(r.URL.Path)
	w.WriteHeader(path.statusCode())
	path.writeBody(w)
}

func (t *TestServer) count(path string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.calls == nil {
		t.calls = make(map[string]int)
	}
	t.calls[path]++
}

// Calls returns how many requests were accepted for each path.
func (t *TestServer) Calls() map[string]int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return maps.Clone(t.calls)
}

// TotalCalls returns how many requests were accepted for all paths.
func (t *TestServer) TotalCalls() int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	var totalCalls int
	for _, call := range t.calls {
		totalCalls += call
	}
	return totalCalls
}

// Path defines how TestServer should reply to a request.
type Path struct {
	// Methods defines which HTTP methods should be accepted. If the request doesn't match, TestServer returns HTTP 405.
	// If Methods is empty, all HTTP methods are accepted.
	Methods []string
	// StatusCode tells TestServer what HTTP status code to return for matching requests.
	// If StatusCode is 0, TestServer returns HTTP 200.
	StatusCode int
	// Body is the body that TestServer will return for the Path. If Body is a string, or a []byte, it is sent as is.
	// Otherwise, it is marshalled as a JSON object.
	Body any
}

func (p Path) validMethod(r *http.Request) bool {
	if len(p.Methods) == 0 {
		return true
	}
	for _, m := range p.Methods {
		if m == r.Method {
			return true
		}
	}
	return false
}

func (p Path) statusCode() int {
	return cmp.Or(p.StatusCode, http.StatusOK)
}

func (p Path) writeBody(w http.ResponseWriter) {
	switch body := p.Body.(type) {
	case string:
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(body))
	case []byte:
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(body)
	default:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}
}
