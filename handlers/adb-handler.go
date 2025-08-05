package handlers

import (
	"adb-server/internal/adb"
	"adb-server/middleware"
	"adb-server/utilities"
	"fmt"
	"net/http"
	"strconv"
)

func HandleListDevices(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the ADB adbClient from the context
	adbClient, ok := middleware.GetADBClient(req)
	if !ok {
		http.Error(res, "ADB client not available", http.StatusInternalServerError)
		return
	}

	devices, deviceListError := adbClient.Devices(req.Context())
	if deviceListError != nil {
		http.Error(res, "error listing devices connected to adb", http.StatusInternalServerError)
		return
	}

	utilities.WriteJSON(res, http.StatusOK, devices)
}

func HandleListPackages(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adbClient, ok := middleware.GetADBClient(req)
	if !ok {
		http.Error(res, "ADB client not available", http.StatusInternalServerError)
		return
	}

	// Get the device-id query parameter
	deviceID := req.URL.Query().Get("device-id")

	// Check if device-id is provided (it's a required parameter)
	if deviceID == "" {
		http.Error(res, "device-id parameter is required", http.StatusBadRequest)
		return
	}

	// Get the include-system parameter and convert to boolean
	includeSystemStr := req.URL.Query().Get("include-system")
	includeSystem := includeSystemStr == "true" || includeSystemStr == "1"

	// Get the uninstalled parameter and convert to boolean
	includeUninstalledStr := req.URL.Query().Get("uninstalled")
	includeUninstalled := includeUninstalledStr == "true" || includeUninstalledStr == "1"

	// Create options for listing packages
	options := adb.ListPackageOptions{
		IncludeSystem:      includeSystem,
		IncludeUninstalled: includeUninstalled,
	}

	// Use the deviceID and options with your ADB client
	packages, err := adbClient.Packages(req.Context(), deviceID, options)

	if err != nil {
		http.Error(res, fmt.Sprintf("error listing packages: %v", err), http.StatusInternalServerError)
		return
	}

	utilities.WriteJSON(res, http.StatusOK, packages)
}

func HandleInstallApp(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adbClient, ok := middleware.GetADBClient(req)
	if !ok {
		http.Error(res, "ADB client not available", http.StatusInternalServerError)
		return
	}

	// Get the device-id query parameter
	deviceID := req.URL.Query().Get("device-id")
	if deviceID == "" {
		http.Error(res, "device-id parameter is required", http.StatusBadRequest)
		return
	}

	// Get the required file path parameter
	filePath := req.URL.Query().Get("path")
	if filePath == "" {
		http.Error(res, "path parameter is required", http.StatusBadRequest)
		return
	}

	// Install the APK using the simplified Install function
	err := adbClient.Install(req.Context(), deviceID, filePath)
	if err != nil {
		http.Error(res, fmt.Sprintf("error installing apk: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	utilities.WriteJSON(res, http.StatusOK, map[string]string{"message": "APK installed successfully"})
}

func HandleUninstallApp(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adbClient, ok := middleware.GetADBClient(req)
	if !ok {
		http.Error(res, "ADB client not available", http.StatusInternalServerError)
		return
	}

	deviceID := req.URL.Query().Get("device-id")
	packageName := req.URL.Query().Get("package")

	if deviceID == "" {
		http.Error(res, "device-id parameter is required", http.StatusBadRequest)
		return
	}

	if packageName == "" {
		http.Error(res, "package parameter is required", http.StatusBadRequest)
		return
	}

	keepDataStr := req.URL.Query().Get("keep-data")
	keepData := keepDataStr == "true" || keepDataStr == "1"

	userStr := req.URL.Query().Get("user")
	user := -1 // Default value (all users)
	if userStr != "" {
		var err error
		user, err = strconv.Atoi(userStr)
		if err != nil {
			http.Error(res, "invalid user parameter", http.StatusBadRequest)
			return
		}
	}

	err := adbClient.Uninstall(req.Context(), deviceID, packageName, keepData, user)
	if err != nil {
		http.Error(res, fmt.Sprintf("error uninstalling package: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	utilities.WriteJSON(res, http.StatusOK, map[string]string{"message": "Package uninstalled successfully"})
}
