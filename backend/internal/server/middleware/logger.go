package middleware

import (
	"log/slog"
	"net/http"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request", slog.String("method", r.Method), slog.String("path", r.RequestURI))
		next.ServeHTTP(w, r)
		slog.Info("Request Result", slog.String("method", r.Method), slog.String("path", r.RequestURI), slog.Any("resp", w.Header()))
	})
}
