package models

import (
	"adb-server/internal/adb"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	Port         int
	ADBClient    adb.Client
	ADBConfig    adb.Config
	HTTPServer   *http.Server
	MainMux      *http.ServeMux
	ProtectedMux *http.ServeMux
	ADBMux       *http.ServeMux
	mu           sync.RWMutex // For thread safety if needed
}

func NewServer(port int) *Server {
	// Initialize ADB config
	adbConfig := adb.Config{
		ADBPath:        "adb",
		ReadTimeout:    30 * time.Second,
		InstallTimeout: 120 * time.Second,
		TempDir:        "",
	}

	// Create ADB client
	adbClient, err := adb.New(adbConfig)
	if err != nil {
		// Handle error appropriately
		panic(err)
	}

	return &Server{
		Port:         port,
		ADBClient:    adbClient,
		ADBConfig:    adbConfig,
		MainMux:      http.NewServeMux(),
		ProtectedMux: http.NewServeMux(),
	}
}

func (server *Server) Start() error {
	server.HTTPServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", server.Port),
		Handler: server.MainMux,
	}

	log.Printf("Server starting on http://%s", server.HTTPServer.Addr)
	log.Printf("Pairing code: %d", server.Port)

	return server.HTTPServer.ListenAndServe()
}
