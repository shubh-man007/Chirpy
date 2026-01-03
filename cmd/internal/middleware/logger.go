package middleware

import (
	"log/slog"
	"net/http"
)

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info(
			"Request",
			"method", r.Method,
			"url", r.URL.Path,
		)

		next.ServeHTTP(w, r)
	})
}

// func LogMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var (
// 			ip     = r.RemoteAddr
// 			method = r.Method
// 			url    = r.URL.String()
// 			proto  = r.Proto
// 		)

// 		userAttrs := slog.Group("user", "ip", ip)
// 		requestAttrs := slog.Group("request", "method", method, "url", url, "proto", proto)

// 		slog.Info("request received", userAttrs, requestAttrs)
// 		next.ServeHTTP(w, r)
// 	})
// }
