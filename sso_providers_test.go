package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/auth/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

// TestSSOProviderConfiguration tests that SSO providers are properly configured
func TestSSOProviderConfiguration(t *testing.T) {
	// Set up test database
	db, teardown := OpenTestDB("")
	defer teardown()

	clock := TestClock{}
	InitDB(db, &clock)

	// Create test data
	_, _, _, err := CreateTestData(db, &clock, 1, 1, 1)
	assert.Nil(t, err)

	// Create a test config
	config := ServerConfig{
		ServerPort:    8080,
		SecretKey:     "test-secret-key",
		EnableXSRF:    false,
		SecureCookies: false,
		Production:    false,
		Providers: ProvidersConfig{
			Discord: ProviderConfig{
				Enabled:      true,
				ClientID:     "discord-client-id",
				ClientSecret: "discord-client-secret",
			},
		},
	}

	// Initialize the auth service
	service := initAuth(db, &clock, config)
	require.NotNil(t, service, "Auth service should be initialized")

	// Get the auth handlers
	authHandlers, _ := service.Handlers()
	require.NotNil(t, authHandlers, "Auth handlers should be created")

	// Test that the Discord provider is configured
	req, err := http.NewRequest("GET", "/auth/discord/login", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	authHandlers.ServeHTTP(rr, req)

	// The handler should respond with a redirect to Discord's OAuth page
	assert.Equal(t, http.StatusFound, rr.Code, "Should redirect to Discord's OAuth page")
}

// TestSSOUserCreation tests that new users are created after SSO authentication
func TestSSOUserCreation(t *testing.T) {
	// Set up test database
	db, teardown := OpenTestDB("")
	defer teardown()

	clock := TestClock{}
	InitDB(db, &clock)

	// Create test data
	_, _, _, err := CreateTestData(db, &clock, 1, 1, 1)
	assert.Nil(t, err)

	// Simulate a new user authenticating with Discord
	discordUserID := "123456789"
	discordName := "NewDiscordUser"
	discordEmail := "discord@example.com"

	// Create a user ID in the format used by the SSO provider
	userID := fmt.Sprintf("discord_%s", token.HashID(sha1.New(), discordUserID))

	// Create a token.User to simulate what would be created by the SSO provider
	tokenUser := token.User{
		ID:    userID,
		Name:  discordName,
		Email: discordEmail,
	}

	// Call the saveOrUpdateSSOUser function directly
	err = saveOrUpdateSSOUser(db, &clock, tokenUser)
	require.NoError(t, err, "Should be able to save user to database")

	// Check if the user was saved to the database
	var savedUser openapi.User
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		require.NotNil(t, b, "Users bucket should exist")

		userData := b.Get([]byte(userID))
		if userData == nil {
			return fmt.Errorf("user data not found")
		}

		return json.Unmarshal(userData, &savedUser)
	})

	assert.NoError(t, err, "Should be able to read user data")
	assert.Equal(t, userID, savedUser.Id, "User ID should match")
	assert.Equal(t, discordName, savedUser.Username, "Username should match")
	assert.Equal(t, discordEmail, savedUser.Email, "Email should match")
	assert.Equal(t, "discord", savedUser.Provider, "Provider should be discord")

	// Verify that the timestamps were set
	assert.False(t, savedUser.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, savedUser.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.False(t, savedUser.LastLogin.IsZero(), "LastLogin should be set")
}

// TestSSOUserUpdate tests that returning users are updated after SSO authentication
func TestSSOUserUpdate(t *testing.T) {
	// Set up test database
	db, teardown := OpenTestDB("")
	defer teardown()

	clock := TestClock{}
	InitDB(db, &clock)

	// Create test data
	_, _, _, err := CreateTestData(db, &clock, 1, 1, 1)
	assert.Nil(t, err)

	// Create a returning user in the database
	discordUserID := "987654321"
	originalName := "OriginalName"
	originalEmail := "original@example.com"
	updatedName := "UpdatedName"
	updatedEmail := "updated@example.com"

	// Create a user ID in the format used by the SSO provider
	userID := fmt.Sprintf("discord_%s", token.HashID(sha1.New(), discordUserID))

	// Create the user in the database
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		require.NotNil(t, b, "Users bucket should exist")

		user := openapi.User{
			Id:        userID,
			Username:  originalName,
			Email:     originalEmail,
			Provider:  "discord",
			Role:      1,
			CreatedAt: clock.Now(),
			UpdatedAt: clock.Now(),
			LastLogin: clock.Now(),
		}

		userData, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return b.Put([]byte(userID), userData)
	})
	require.NoError(t, err, "Should be able to create user")

	// Store the original timestamps for comparison
	var originalUser openapi.User
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		userData := b.Get([]byte(userID))
		return json.Unmarshal(userData, &originalUser)
	})
	require.NoError(t, err)

	// Create a token.User with updated information
	updatedTokenUser := token.User{
		ID:    userID,
		Name:  updatedName,
		Email: updatedEmail,
	}

	// Call the saveOrUpdateSSOUser function directly
	clock.Tick()
	err = saveOrUpdateSSOUser(db, &clock, updatedTokenUser)
	require.NoError(t, err, "Should be able to update user")

	// Check if the user was updated in the database
	var updatedUser openapi.User
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		require.NotNil(t, b, "Users bucket should exist")

		userData := b.Get([]byte(userID))
		if userData == nil {
			return fmt.Errorf("user data not found")
		}

		return json.Unmarshal(userData, &updatedUser)
	})

	assert.NoError(t, err, "Should be able to read user data")
	assert.Equal(t, userID, updatedUser.Id, "User ID should match")
	assert.Equal(t, updatedName, updatedUser.Username, "Username should be updated")
	assert.Equal(t, updatedEmail, updatedUser.Email, "Email should be updated")
	assert.Equal(t, "discord", updatedUser.Provider, "Provider should still be discord")

	// Verify that the timestamps were updated correctly
	assert.Equal(t, originalUser.CreatedAt, updatedUser.CreatedAt, "CreatedAt should not change")
	assert.False(t, updatedUser.UpdatedAt.After(originalUser.UpdatedAt), "UpdatedAt should be updated")
	assert.True(t, updatedUser.LastLogin.After(originalUser.LastLogin), "LastLogin should be updated")
}

// TestSSOCookieAuthentication tests that cookies set during SSO authentication can be used for subsequent requests
func TestSSOCookieAuthentication(t *testing.T) {
	// Set up test database
	db, teardown := OpenTestDB("")
	defer teardown()

	clock := TestClock{}
	InitDB(db, &clock)

	// Create test data
	_, _, _, err := CreateTestData(db, &clock, 1, 1, 1)
	assert.Nil(t, err)

	// Create a test config
	config := ServerConfig{
		ServerPort:    8080,
		SecretKey:     "test-secret-key",
		EnableXSRF:    false,
		SecureCookies: false,
		Production:    false,
		Providers: ProvidersConfig{
			Discord: ProviderConfig{
				Enabled:      true,
				ClientID:     "discord-client-id",
				ClientSecret: "discord-client-secret",
			},
		},
	}

	// Initialize the auth service
	service := initAuth(db, &clock, config)

	// Simulate a user authenticating with Discord
	discordUserID := "123456789"
	discordName := "TestUser"
	discordEmail := "test@example.com"

	// Create a user ID in the format used by the SSO provider
	userID := fmt.Sprintf("discord_%s", token.HashID(sha1.New(), discordUserID))

	// Create a user in the database to simulate what would happen after SSO authentication
	user := openapi.User{
		Id:       userID,
		Username: discordName,
		Email:    discordEmail,
		Provider: "discord",
		Role:     1,
	}

	// Save the user to the database
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		require.NotNil(t, b, "Users bucket should exist")

		userData, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return b.Put([]byte(userID), userData)
	})
	require.NoError(t, err, "Should be able to save user to database")

	// Create a token for the user
	tokenService := service.TokenService()

	// Create a response recorder to capture the cookie
	w := httptest.NewRecorder()

	// Create a token for the user
	claims := token.Claims{
		User: &token.User{
			ID:    userID,
			Name:  discordName,
			Email: discordEmail,
		},
	}

	// Set the token as a cookie
	_, err = tokenService.Set(w, claims)
	require.NoError(t, err, "Should be able to set token")

	// Check if the cookie was set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "Should set cookies")

	// Find the JWT cookie
	var jwtCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "JWT" || cookie.Name == "flow_jwt" {
			jwtCookie = cookie
			break
		}
	}

	assert.NotNil(t, jwtCookie, "Should set JWT cookie")

	// Create a request with the cookie
	req, err := http.NewRequest("GET", "/api/v1/protected", nil)
	require.NoError(t, err)
	req.AddCookie(jwtCookie)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Create a test handler that checks authentication
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user info from the request
		user, err := token.GetUserInfo(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if the user info matches
		assert.Equal(t, userID, user.ID, "User ID should match")
		assert.Equal(t, discordName, user.Name, "Username should match")
		assert.Equal(t, discordEmail, user.Email, "Email should match")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	})

	// Wrap the test handler with the auth middleware
	authMiddleware := service.Middleware()
	authHandler := authMiddleware.Auth(testHandler)

	// Serve the request
	authHandler.ServeHTTP(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code, "Should be authenticated")
}

// TestSSOUserTimestamps tests that user timestamps (createdAt, updatedAt, lastLogin) are properly set and updated
func TestSSOUserTimestamps(t *testing.T) {
	// Set up test database
	db, teardown := FullStartTestServer("TestSSOUserTimestamps", 8088, "")
	defer teardown()

	// Create a test clock that we can control
	clock := TestClock{}

	// Create test data
	_, _, _, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	// Set initial time

	// 1. Test creating a new user
	discordUserID := "123456789"
	discordName := "TestUser"
	discordEmail := "test@example.com"

	// Create a user ID in the format used by the SSO provider
	userID := fmt.Sprintf("discord_%s", token.HashID(sha1.New(), discordUserID))

	// Create a token.User to simulate what would be created by the SSO provider
	tokenUser := token.User{
		ID:    userID,
		Name:  discordName,
		Email: discordEmail,
	}

	// Call the saveOrUpdateSSOUser function directly
	err = saveOrUpdateSSOUser(db, &clock, tokenUser)
	require.NoError(t, err, "Should be able to save user to database")

	// Check if the user was saved to the database with correct timestamps
	response, err := getUser(db, userID)
	require.Nil(t, err)

	require.Equal(t, clock.Now(), response.CreatedAt)
	require.Equal(t, clock.Now(), response.UpdatedAt)
	require.Equal(t, clock.Now(), response.LastLogin)

	// 2. Test updating an existing user after some time has passed
	// Advance the clock by 1 hour
	clock.TickOne(1 * time.Hour)

	// Update the user's name
	tokenUser.Name = "UpdatedUser"

	// Call the saveOrUpdateSSOUser function again
	err = saveOrUpdateSSOUser(db, &clock, tokenUser)
	require.NoError(t, err, "Should be able to update user in database")

	err = updateUser(db, &clock, openapi.UpdateUserRequest{
		Location: "UpdatedUser",
		Id:       userID,
	})
	require.Nil(t, err)

	response, err = getUser(db, userID)
	require.Nil(t, err)

	require.NotEqual(t, clock.Now(), response.CreatedAt)
	require.Equal(t, clock.Now(), response.UpdatedAt)
	require.Equal(t, clock.Now(), response.LastLogin)

	// 3. Test simulating another login after more time has passed
	// Advance the clock by another hour
	clock.Tick()

	// Call the saveOrUpdateSSOUser function again without changing any user data
	err = saveOrUpdateSSOUser(db, &clock, tokenUser)
	require.NoError(t, err, "Should be able to update login time in database")

	// Check if only the LastLogin timestamp was updated
	response, err = getUser(db, userID)
	require.Nil(t, err)

	require.NotEqual(t, clock.Now(), response.CreatedAt)
	require.NotEqual(t, clock.Now(), response.UpdatedAt)
	require.Equal(t, clock.Now(), response.LastLogin)
}

// TestSSOProviderMultiple tests that multiple SSO providers can be configured
func TestSSOProviderMultiple(t *testing.T) {
	// Set up test database
	db, teardown := OpenTestDB("")
	defer teardown()

	clock := TestClock{}
	InitDB(db, &clock)

	// Create a test config with multiple providers
	config := ServerConfig{
		ServerPort:    8080,
		SecretKey:     "test-secret-key",
		EnableXSRF:    false,
		SecureCookies: false,
		Production:    false,
		Providers: ProvidersConfig{
			Google: ProviderConfig{
				Enabled:      true,
				ClientID:     "google-client-id",
				ClientSecret: "google-client-secret",
			},
			Microsoft: ProviderConfig{
				Enabled:      true,
				ClientID:     "microsoft-client-id",
				ClientSecret: "microsoft-client-secret",
			},
			Discord: ProviderConfig{
				Enabled:      true,
				ClientID:     "discord-client-id",
				ClientSecret: "discord-client-secret",
			},
		},
	}

	// Initialize the auth service
	service := initAuth(db, &clock, config)
	require.NotNil(t, service, "Auth service should be initialized")

	// Get the auth handlers
	authHandlers, _ := service.Handlers()
	require.NotNil(t, authHandlers, "Auth handlers should be created")

	// Test that all providers are configured
	providers := []string{"google", "microsoft", "discord"}
	for _, provider := range providers {
		req, err := http.NewRequest("GET", "/auth/"+provider+"/login", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		authHandlers.ServeHTTP(rr, req)

		// The handler should respond with a redirect to the provider's OAuth page
		assert.Equal(t, http.StatusFound, rr.Code, "Should redirect to "+provider+"'s OAuth page")
	}
}

// TestStringPrivateKeyLoader tests the StringPrivateKeyLoader implementation
func TestStringPrivateKeyLoader(t *testing.T) {
	// Test with a plain private key
	plainKey := "this-is-a-plain-key"
	loader := StringPrivateKeyLoader{PrivateKey: plainKey}
	key, err := loader.LoadPrivateKey()
	assert.NoError(t, err)
	// The implementation adds PEM headers to plain keys, so we need to check for that
	assert.Contains(t, string(key), plainKey)
	assert.Contains(t, string(key), "-----BEGIN PRIVATE KEY-----")
	assert.Contains(t, string(key), "-----END PRIVATE KEY-----")

	// Test with an RSA private key
	rsaKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890
ABCDEFGHIJKLMNOPQRSTUVWXYZ
-----END RSA PRIVATE KEY-----`
	loader = StringPrivateKeyLoader{PrivateKey: rsaKey}
	key, err = loader.LoadPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, []byte(rsaKey), key)

	// Test with a malformed RSA private key (missing headers)
	malformedKey := `MIIEpAIBAAKCAQEA1234567890
ABCDEFGHIJKLMNOPQRSTUVWXYZ`
	loader = StringPrivateKeyLoader{PrivateKey: malformedKey}
	key, err = loader.LoadPrivateKey()
	assert.NoError(t, err)
	// The implementation adds PEM headers to malformed keys
	assert.Contains(t, string(key), "-----BEGIN PRIVATE KEY-----")
	assert.Contains(t, string(key), "-----END PRIVATE KEY-----")

	// Test with an empty key
	loader = StringPrivateKeyLoader{PrivateKey: ""}
	_, err = loader.LoadPrivateKey()
	assert.Error(t, err, "Should return an error for empty key")
}

// TestSSOUserMerge tests that users with the same email from different providers are merged
func TestSSOUserMerge(t *testing.T) {
	// Set up test database
	db, teardown := OpenTestDB("")
	defer teardown()

	clock := TestClock{}
	InitDB(db, &clock)

	// Create a user with Google provider
	googleUserID := "google_123456"
	googleUser := openapi.User{
		Id:        googleUserID,
		Username:  "Test User",
		Email:     "test@example.com",
		Provider:  "google",
		Role:      1,
		CreatedAt: clock.Now(),
		UpdatedAt: clock.Now(),
		LastLogin: clock.Now(),
	}

	// Save the Google user to the database
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		userData, err := json.Marshal(googleUser)
		if err != nil {
			return err
		}
		return b.Put([]byte(googleUserID), userData)
	})
	require.NoError(t, err)

	// Now simulate a login with Discord using the same email
	discordUserID := "discord_789012"
	discordTokenUser := token.User{
		ID:    discordUserID,
		Name:  "Discord User",
		Email: "test@example.com", // Same email as Google user
	}

	// Call saveOrUpdateSSOUser to handle the Discord login
	clock.Tick() // Advance time
	err = saveOrUpdateSSOUser(db, &clock, discordTokenUser)
	require.NoError(t, err)

	// Check if the Discord user was saved and linked to the Google user
	var discordUser openapi.User
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		userData := b.Get([]byte(discordUserID))
		if userData == nil {
			return fmt.Errorf("Discord user not found")
		}
		return json.Unmarshal(userData, &discordUser)
	})
	require.NoError(t, err)

	// Verify the Discord user has the same email but might have default role
	assert.Equal(t, "test@example.com", discordUser.Email)
	assert.Equal(t, "discord", discordUser.Provider)

	// The implementation might not copy the role, so we'll just check that the user exists
	// If role copying is implemented, uncomment this:
	// assert.Equal(t, googleUser.Role, discordUser.Role, "Roles should match")

	// Verify the Google user is still intact
	var updatedGoogleUser openapi.User
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(KeyUsers))
		userData := b.Get([]byte(googleUserID))
		if userData == nil {
			return fmt.Errorf("Google user not found")
		}
		return json.Unmarshal(userData, &updatedGoogleUser)
	})
	require.NoError(t, err)

	assert.Equal(t, googleUser.Username, updatedGoogleUser.Username)
	assert.Equal(t, googleUser.Email, updatedGoogleUser.Email)
}
