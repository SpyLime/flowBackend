package main

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/provider"
	"github.com/go-pkgz/auth/token"
	"go.etcd.io/bbolt"
	"golang.org/x/oauth2"
)

// StringPrivateKeyLoader implements the PrivateKeyLoaderInterface for loading a private key from a string
type StringPrivateKeyLoader struct {
	PrivateKey string
}

// LoadPrivateKey implements the PrivateKeyLoaderInterface method
func (s *StringPrivateKeyLoader) LoadPrivateKey() ([]byte, error) {
	if s.PrivateKey == "" {
		return nil, fmt.Errorf("empty private key not allowed")
	}

	// Check if the private key is already in PEM format
	if !strings.HasPrefix(s.PrivateKey, "-----BEGIN") {
		// Try to determine the key type
		keyType := "PRIVATE KEY"
		if strings.Contains(strings.ToLower(s.PrivateKey), "begin ec private key") {
			keyType = "EC PRIVATE KEY"
		} else if strings.Contains(strings.ToLower(s.PrivateKey), "begin rsa private key") {
			keyType = "RSA PRIVATE KEY"
		}

		// Clean the key content (remove any headers/footers if present)
		cleanKey := strings.ReplaceAll(s.PrivateKey, "-----BEGIN "+keyType+"-----", "")
		cleanKey = strings.ReplaceAll(cleanKey, "-----END "+keyType+"-----", "")
		cleanKey = strings.TrimSpace(cleanKey)

		// Format the key properly with headers and line breaks
		pemKey := fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----",
			keyType, cleanKey, keyType)

		return []byte(pemKey), nil
	}

	return []byte(s.PrivateKey), nil
}

// ConfigureSSO adds all the SSO providers to the auth service
func ConfigureSSO(service *auth.Service, config ServerConfig, serverPort int, db *bbolt.DB) {
	fmt.Println("Configuring SSO providers...")

	// Add Google provider if enabled
	if config.Providers.Google.Enabled {
		fmt.Println("Adding Google provider...")
		service.AddProvider("google",
			config.Providers.Google.ClientID,
			config.Providers.Google.ClientSecret)
		fmt.Println("Google provider added successfully")
	}

	// Add Microsoft provider if enabled
	if config.Providers.Microsoft.Enabled {
		service.AddProvider("microsoft",
			config.Providers.Microsoft.ClientID,
			config.Providers.Microsoft.ClientSecret)
	}

	// Add Facebook provider if enabled
	if config.Providers.Facebook.Enabled {
		service.AddProvider("facebook",
			config.Providers.Facebook.ClientID,
			config.Providers.Facebook.ClientSecret)
	}

	// Add Discord provider if enabled
	if config.Providers.Discord.Enabled {
		fmt.Println("Adding Discord provider...")
		fmt.Printf("Discord client ID: %s\n", config.Providers.Discord.ClientID)
		fmt.Printf("Discord client secret: %s\n", config.Providers.Discord.ClientSecret[:5]+"...")

		// Create a custom Discord OAuth2 provider with explicit redirect URI
		// This ensures the correct redirect URI is used regardless of the production flag
		callbackURL := fmt.Sprintf("http://localhost:%d/auth/discord/callback", serverPort)
		fmt.Printf("Discord callback URL: %s\n", callbackURL)

		// We'll use a custom HTML page to handle the redirect
		// The callback.html page will parse the user data and redirect to the topics page

		// Add Discord provider with custom options
		// Create a custom handler for Discord
		discordHandler := provider.CustomHandlerOpt{
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
			InfoURL: "https://discord.com/api/users/@me",
			Scopes:  []string{"identify", "email"},
			// Note: We can't add a custom success handler here
			// We'll need to modify the auth service to handle the redirect
			// Add a custom MapUserFn to handle Discord's user data format
			MapUserFn: func(data provider.UserData, _ []byte) token.User {
				fmt.Printf("Discord user data: %+v\n", data)

				// Create a user from the Discord data
				user := token.User{
					ID:   "discord_" + token.HashID(sha1.New(), data.Value("id")),
					Name: data.Value("username"),
				}

				// Add avatar URL if available
				if avatar := data.Value("avatar"); avatar != "" {
					user.Picture = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", data.Value("id"), avatar)
				}

				// Add email if available
				if email := data.Value("email"); email != "" {
					user.Email = email
				}

				// Save the user to the database
				err := saveOrUpdateSSOUser(db, user)
				if err != nil {
					fmt.Printf("Error saving Discord user to database: %v\n", err)
				}

				return user
			},
		}

		// Add the Discord provider with the custom handler
		service.AddCustomProvider("discord",
			auth.Client{
				Cid:     config.Providers.Discord.ClientID,
				Csecret: config.Providers.Discord.ClientSecret,
			},
			discordHandler,
		)

		fmt.Println("Discord provider added successfully with custom configuration")
	}

	// Add Twitter provider if enabled
	if config.Providers.Twitter.Enabled {
		service.AddProvider("twitter",
			config.Providers.Twitter.ClientID,
			config.Providers.Twitter.ClientSecret)
	}

	// Add Apple provider if enabled
	if config.Providers.Apple.Enabled {
		// Skip if required fields are missing
		if config.Providers.Apple.ClientID == "" || config.Providers.Apple.TeamID == "" ||
			config.Providers.Apple.KeyID == "" || config.Providers.Apple.ClientSecret == "" {
			fmt.Println("Skipping Apple provider - missing required configuration")
		} else {
			// Apple requires a special setup with team ID and key ID
			appleConfig := provider.AppleConfig{
				ClientID: config.Providers.Apple.ClientID,
				TeamID:   config.Providers.Apple.TeamID,
				KeyID:    config.Providers.Apple.KeyID,
			}

			// Create a private key loader that loads the key from a string
			privateKeyLoader := &StringPrivateKeyLoader{PrivateKey: config.Providers.Apple.ClientSecret}

			// Add the Apple provider
			if err := service.AddAppleProvider(appleConfig, privateKeyLoader); err != nil {
				fmt.Printf("Failed to add Apple provider: %v\n", err)
			} else {
				fmt.Println("Successfully added Apple provider")
			}
		}
	}
}
