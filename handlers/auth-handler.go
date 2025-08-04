package handlers

import (
	"adb-server/authentication"
	"adb-server/models"
	"adb-server/utilities"
	"net/http"
)

var authTokenLength int = 32

// I know this isnt the most secure function
// But it is quick enough to implement and secure enough.
// We change the port the server is running on every 60 seconds. ( I know you could still brute force it but lowers the chances )
// The server closes after 180 seconds of no pairing
func PairWithServer(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != http.MethodPost {
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only allow one pairing per session
	if authentication.IsPaired() {
		http.Error(responseWriter, "a client is already paired", http.StatusConflict)
		return
	}

	serverAuthenticationToken, tokenError := authentication.GenerateAuthToken(authTokenLength)

	if tokenError != nil {
		http.Error(responseWriter, "problem generating auth token", http.StatusInternalServerError)
		return
	}

	// Create a new cookie with the authentication token
	cookie := &http.Cookie{
		Name:     "X-Auth-Token",
		Value:    serverAuthenticationToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(responseWriter, cookie)

	authenticationResponse := models.PairingResponse{
		Status: "Success",
	}

	utilities.WriteJSON(responseWriter, 200, authenticationResponse)
}
