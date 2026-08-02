package middleware

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const RATELIMIT = time.Minute / 5

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s, %s, %s", req.Method, req.RequestURI, time.Since(start))
	})
}

// Recovery is a middleware that recovers from panics and returns a 500 error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Println(string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, req)
	})
}

// Rate limiting middleware
// - Limits the number of requests per minute to certain requests per minute
func RateLimit(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if limiter.Allow() != true {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				// log.Println(string(debug.Stack()))
			}
			next.ServeHTTP(w, req)
		})
	}
}

// Authentication middleware
func Authenticate(validate func(token string) error) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			authorization := req.Header.Get("Authorization")
			scheme, token, ok := strings.Cut(authorization, " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") {
				http.Error(w, "invalid authorization format!", http.StatusUnauthorized)
				return
			}
			if err := validate(token); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(req.Context(), "auth-token", token)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}
