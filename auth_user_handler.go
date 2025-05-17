package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-pkgz/auth/token"
)

// AuthUserHandler handles the /auth/user endpoint
func AuthUserHandler(w http.ResponseWriter, r *http.Request) {
	// AuthUserHandler called

	// Get user info from the request
	user, err := token.GetUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// User info retrieved successfully

	// Return user info as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
