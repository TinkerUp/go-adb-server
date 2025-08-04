package utilities

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
)

func WriteJSON(responseWriter http.ResponseWriter, status int, jsonContent any) {
	responseWriter.Header().Set("Content-Type", "application/json")

	responseWriter.WriteHeader(status)

	_ = json.NewEncoder(responseWriter).Encode(jsonContent) // Ignoring JSON write errors for now
}

// pickRandomPort tries to bind to an available random port in [min,max].
func PickRandomPort(min int, max int) int {
	for i := 0; i < 20; i++ {
		port := rand.Intn(max-min+1) + min

		listener, error := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))

		if error == nil {
			_ = listener.Close()
			return port
		}
	}

	// Fallback to 0 (let OS choose), then extract the port
	listener, error := net.Listen("tcp", "127.0.0.1:0")

	if error != nil {
		log.Fatal("failed to find a free port:", error)
	}

	defer listener.Close()

	_, portString, _ := net.SplitHostPort(listener.Addr().String()) // Extracting only the port, discarding others

	var port int

	fmt.Sscanf(portString, "%d", &port) // converting port to int

	return port
}
