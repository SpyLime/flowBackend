package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-pkgz/auth/token"
)

// AuthUserHandler handles the /auth/user endpoint
func AuthUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("AuthUserHandler called with URL: %s\n", r.URL.String())
	fmt.Printf("Request method: %s\n", r.Method)
	fmt.Printf("Request headers: %v\n", r.Header)

	// Get user info from the request
	user, err := token.GetUserInfo(r)
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Printf("User info retrieved successfully: %+v\n", user)

	// Return user info as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
