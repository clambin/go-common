package httpserver

import (
	"github.com/clambin/go-common/set"
	"net/http"
)

// MethodFilter only passes on the request if the http request's methos matches once of the methods.
// If no methods are provided, MethodFilter defaults to http.MethodGet.
func MethodFilter(methods ...string) func(next http.Handler) http.Handler {
	if len(methods) == 0 {
		methods = []string{http.MethodGet}
	}
	methodSet := set.Create(methods...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !methodSet.Contains(req.Method) {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}
