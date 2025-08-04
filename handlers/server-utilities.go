package handlers

import (
	"net/http"
	"time"

	"adb-server/models"
	"adb-server/utilities"
)

// handleHealth is a handler function for the health check endpoint.
func HandleServerHealth(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != http.MethodGet {
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	healthResponse := models.HealthResponse{
		Time:   time.Now().Format(time.RFC3339),
		Status: "ok",
	}

	utilities.WriteJSON(responseWriter, http.StatusOK, healthResponse)
}
