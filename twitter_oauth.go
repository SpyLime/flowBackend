package main

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/lgr"
	"github.com/golang-jwt/jwt"
	"go.etcd.io/bbolt"
)

// TwitterOAuthHandler handles Twitter OAuth 2.0 with PKCE
func TwitterOAuthHandler(w http.ResponseWriter, r *http.Request, config ServerConfig) {
	// Generate code verifier (random string between 43-128 chars)
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		http.Error(w, "Failed to generate code verifier", http.StatusInternalServerError)
		return
	}

	// Store code verifier in cookie for later verification
	http.SetCookie(w, &http.Cookie{
		Name:     "twitter_code_verifier",
		Value:    codeVerifier,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.SecureCookies,
		MaxAge:   int(10 * time.Minute.Seconds()),
	})

	// Generate code challenge (SHA256 hash of verifier, base64url encoded)
	codeChallenge := generateCodeChallenge(codeVerifier)

	var redirectHost string
	redirectHost = fmt.Sprintf("http://localhost:%d", config.ServerPort)
	if config.Server {
		if config.Production {
			redirectHost = "https://flow.ubuck.org"
		} else {
			redirectHost = "https://flow-test.ubuck.org"
		}
	}

	redirectURL := fmt.Sprintf("%s/auth/twitter/callback", redirectHost)

	// Build authorization URL
	authURL := fmt.Sprintf(
		"https://twitter.com/i/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
		url.QueryEscape(config.Providers.Twitter.ClientID),
		url.QueryEscape(redirectURL),
		url.QueryEscape("users.read tweet.read"),
		url.QueryEscape("twitter_state"), // In production, use a secure random state
		url.QueryEscape(codeChallenge),
	)

	// Redirect to Twitter authorization page
	http.Redirect(w, r, authURL, http.StatusFound)
}

// TwitterCallbackHandler handles the OAuth callback from Twitter
func TwitterCallbackHandler(w http.ResponseWriter, r *http.Request, config ServerConfig, db *bbolt.DB, clock Clock) {
	// Get the authorization code from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Get the code verifier from the cookie
	cookie, err := r.Cookie("twitter_code_verifier")
	if err != nil {
		lgr.Printf("ERROR: Code verifier cookie not found: %v", err)
		http.Error(w, "Code verifier not found", http.StatusBadRequest)
		return
	}
	codeVerifier := cookie.Value

	var redirectHost string
	redirectHost = fmt.Sprintf("http://localhost:%d", config.ServerPort)
	if config.Server {
		if config.Production {
			redirectHost = "https://flow.ubuck.org"
		} else {
			redirectHost = "https://flow-test.ubuck.org"
		}
	}

	redirectURL := fmt.Sprintf("%s/auth/twitter/callback", redirectHost)

	// Exchange the authorization code for an access token
	tokenURL := "https://api.twitter.com/2/oauth2/token"
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", fmt.Sprintf(redirectURL))
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		lgr.Printf("ERROR: Failed to create token request: %v", err)
		http.Error(w, "Failed to create token request", http.StatusInternalServerError)
		return
	}

	// Add Basic Authentication header for client credentials
	req.SetBasicAuth(config.Providers.Twitter.ClientID, config.Providers.Twitter.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		lgr.Printf("ERROR: Failed to exchange token: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		lgr.Printf("ERROR: Token response error: %s - %s", resp.Status, string(respBody))
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Parse the token response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		lgr.Printf("ERROR: Failed to parse token response: %v", err)
		http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
		return
	}

	// Use the access token to get user information
	userInfoURL := "https://api.twitter.com/2/users/me?user.fields=profile_image_url,name,username,id"
	userReq, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		lgr.Printf("ERROR: Failed to create user info request: %v", err)
		http.Error(w, "Failed to create user info request", http.StatusInternalServerError)
		return
	}
	userReq.Header.Add("Authorization", "Bearer "+tokenResponse.AccessToken)

	userResp, err := client.Do(userReq)
	if err != nil {
		lgr.Printf("ERROR: Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(userResp.Body)
		lgr.Printf("ERROR: User info response error: %s - %s", userResp.Status, string(respBody))
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Parse the user info response
	var userResponse struct {
		Data struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Username        string `json:"username"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userResponse); err != nil {
		lgr.Printf("ERROR: Failed to parse user info: %v", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Create a user object for our system
	user := token.User{
		ID:      "twitter_" + token.HashID(sha1.New(), userResponse.Data.ID),
		Name:    userResponse.Data.Name,
		Email:   userResponse.Data.Username + "@twitter.com", // Twitter doesn't provide email
		Picture: userResponse.Data.ProfileImageURL,
	}

	// Save the user to the database
	err = saveOrUpdateSSOUser(db, clock, user)
	if err != nil {
		lgr.Printf("ERROR: Failed to save user: %v", err)
		// Continue anyway, as this is not critical
	}

	// Create a JWT token for the user using the existing function
	tokenString, err := createUserToken(user, config.SecretKey)
	if err != nil {
		lgr.Printf("ERROR: Failed to create JWT token: %v", err)
		http.Error(w, "Failed to create user session", http.StatusInternalServerError)
		return
	}

	// Set the token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "flow_jwt",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.SecureCookies,
		MaxAge:   int(24 * time.Hour.Seconds()),
	})

	// Redirect directly to the topics page
	frontendURL := "http://localhost:3000"
	// For production, use the actual domain
	if config.Server {
		if config.Production {
			frontendURL = "https://flow.ubuck.org"
		} else {
			frontendURL = "https://flow-test.ubuck.org"
		}
	}
	redirectURL = frontendURL + "/topics"
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// Helper functions for PKCE
func generateCodeVerifier() (string, error) {
	// Generate a random string between 43-128 characters
	b := make([]byte, 64) // 64 bytes = 86 characters in base64
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64URLEncode(b), nil
}

func generateCodeChallenge(verifier string) string {
	// SHA256 hash of the verifier
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64URLEncode(h.Sum(nil))
}

func base64URLEncode(data []byte) string {
	// Base64 URL encoding without padding
	encoded := base64.StdEncoding.EncodeToString(data)
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	encoded = strings.ReplaceAll(encoded, "=", "")
	return encoded
}

// Helper function to create a JWT token for a user
func createUserToken(user token.User, secretKey string) (string, error) {
	// Create a new token object
	claims := token.Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Issuer:    "FL",
		},
		User: &user,
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	return token.SignedString([]byte(secretKey))
}
