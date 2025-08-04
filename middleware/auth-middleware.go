package middleware

import (
	"adb-server/authentication"
	"net/http"
)

func ProtectedRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			// Get the cookie from the request
			cookie, err := req.Cookie("X-Auth-Token")

			// If the cookie is not found or is empty, deny access
			if err != nil || cookie.Value == "" {
				http.Error(res, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if the token is valid
			if !authentication.VerifyAuthToken(cookie.Value) {
				http.Error(res, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// If everything is okay, call the next handler in the chain
			next.ServeHTTP(res, req)
		},
	)
}
