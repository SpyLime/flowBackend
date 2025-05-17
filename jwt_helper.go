package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// JWTPayload represents the payload of a JWT token
type JWTPayload struct {
	User struct {
		Name    string `json:"name"`
		ID      string `json:"id"`
		Picture string `json:"picture"`
		Email   string `json:"email"`
	} `json:"user"`
	Exp int64  `json:"exp"`
	Jti string `json:"jti"`
	Iat int64  `json:"iat"`
	Iss string `json:"iss"`
}

// ExtractUserFromJWT extracts user information from a JWT token without verifying the signature
func ExtractUserFromJWT(tokenString string) (*JWTPayload, error) {
	// Split the token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %v", err)
	}

	// Parse the payload
	var jwtPayload JWTPayload
	err = json.Unmarshal(payload, &jwtPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payload: %v", err)
	}

	// JWT payload extracted successfully

	return &jwtPayload, nil
}
