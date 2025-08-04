package main

import (
	"adb-server/handlers"
	"adb-server/middleware"
	"adb-server/utilities"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := utilities.PickRandomPort(35000, 49151) // Choosing Random port for basic security

	mainMux := http.NewServeMux()      // Handles all routes
	protectedMux := http.NewServeMux() // For all routes which require authentication. For now, all routes require authentication

	protectedMux.HandleFunc("/v1/health", handlers.HandleServerHealth)

	protectedRouteHandler := middleware.ProtectedRoute(protectedMux) // All routes

	mainMux.HandleFunc("/v1/pair", handlers.PairWithServer)
	mainMux.Handle("/v1/", protectedRouteHandler)

	serverAddress := fmt.Sprintf("127.0.0.1:%d", port)

	log.Printf("Server starting on http://%s", serverAddress)
	log.Printf("Pairing code: %d", port)

	if err := http.ListenAndServe(serverAddress, mainMux); err != nil { // Not gonna lie, this is a pretty sick feature of go.
		log.Fatal(err)
	}
}
