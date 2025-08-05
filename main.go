package main

import (
	"adb-server/handlers"
	"adb-server/middleware"
	"adb-server/models"
	"adb-server/utilities"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := utilities.PickRandomPort(35000, 49151)

	server := models.NewServer(port)

	// Set up protected routes
	server.ProtectedMux.HandleFunc("/v1/health", handlers.HandleServerHealth)
	server.ProtectedMux.HandleFunc("/v1/adb/list-devices", handlers.HandleListDevices)
	server.ProtectedMux.HandleFunc("/v1/adb/list-packages", handlers.HandleListPackages)
	server.ProtectedMux.HandleFunc("/v1/adb/install-package", handlers.HandleInstallApp)
	server.ProtectedMux.HandleFunc("/v1/adb/uninstall-package", handlers.HandleUninstallApp)

	protectedRouteHandler := middleware.ProtectedRoute(server.ProtectedMux)

	// Applying ADB client middleware it to all protected routes since ADB operations would be protected
	// Must change in the future though
	adbRouteHandler := middleware.WithADBClient(server.ADBClient)(protectedRouteHandler)

	server.MainMux.HandleFunc("/v1/pair", handlers.PairWithServer)
	server.MainMux.Handle("/v1/", adbRouteHandler)

	serverAddress := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("Server starting on http://%s", serverAddress)
	log.Printf("Pairing code: %d", port)

	if err := http.ListenAndServe(serverAddress, server.MainMux); err != nil {
		log.Fatal(err)
	}
}
