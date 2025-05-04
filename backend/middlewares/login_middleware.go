package middlewares

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// LoggingMiddleware logs basic request info.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("Incoming request: %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}