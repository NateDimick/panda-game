package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("Request", zap.String("method", r.Method), zap.String("path", r.RequestURI))
		next.ServeHTTP(w, r)
	})
}
