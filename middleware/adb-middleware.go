package middleware

import (
	"adb-server/internal/adb"
	"context"
	"net/http"
)

type contextKey string

const adbClientKey contextKey = "adbClient"

func WithADBClient(client adb.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), adbClientKey, client)

				r = r.WithContext(ctx)

				next.ServeHTTP(w, r)
			},
		)
	}
}

func GetADBClient(r *http.Request) (adb.Client, bool) {
	client, ok := r.Context().Value(adbClientKey).(adb.Client)
	return client, ok
}
