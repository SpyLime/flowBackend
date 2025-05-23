package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/provider"
	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/lgr"
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
func ConfigureSSO(service *auth.Service, config ServerConfig, serverPort int, db *bbolt.DB, clock Clock) {
	// Configure SSO providers
	if config.Providers.Google.Enabled {

		googleHandler := provider.CustomHandlerOpt{
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			InfoURL: "https://www.googleapis.com/oauth2/v3/userinfo",
			Scopes:  []string{"openid", "profile", "email"},
			MapUserFn: func(data provider.UserData, _ []byte) token.User {
				user := token.User{
					ID:      "google_" + token.HashID(sha1.New(), data.Value("sub")),
					Name:    data.Value("name"),
					Email:   data.Value("email"),
					Picture: data.Value("picture"),
				}
				// persist to your BoltDB
				_ = saveOrUpdateSSOUser(db, clock, user)
				return user
			},
		}

		service.AddCustomProvider("google",
			auth.Client{Cid: config.Providers.Google.ClientID, Csecret: config.Providers.Google.ClientSecret},
			googleHandler,
		)
	}

	// Add Microsoft provider if enabled
	if config.Providers.Microsoft.Enabled {
		// Create a custom Microsoft OAuth2 provider
		microsoftHandler := provider.CustomHandlerOpt{
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			},
			InfoURL: "https://graph.microsoft.com/v1.0/me",
			// Request profile, email, and photo permissions
			Scopes: []string{"openid", "profile", "email", "User.Read"},
			MapUserFn: func(data provider.UserData, _ []byte) token.User {
				// Create a user from the Microsoft data
				user := token.User{
					ID:   "microsoft_" + token.HashID(sha1.New(), data.Value("id")),
					Name: data.Value("displayName"),
				}

				// Add email if available
				if email := data.Value("mail"); email != "" {
					user.Email = email
				} else if upn := data.Value("userPrincipalName"); upn != "" {
					// Fallback to userPrincipalName which is often the email
					user.Email = upn
				}

				// Microsoft Graph API doesn't return the photo in the user info
				// We would need to make a separate request to get the photo
				// For now, we'll leave the picture empty
				// A more complete implementation would make a request to:
				// https://graph.microsoft.com/v1.0/me/photo/$value

				// Save the user to the database
				_ = saveOrUpdateSSOUser(db, clock, user)

				return user
			},
		}

		// Add the Microsoft provider with the custom handler
		service.AddCustomProvider("microsoft",
			auth.Client{
				Cid:     config.Providers.Microsoft.ClientID,
				Csecret: config.Providers.Microsoft.ClientSecret,
			},
			microsoftHandler,
		)
	}

	// Add Facebook provider if enabled
	if config.Providers.Facebook.Enabled {
		// Create a custom Facebook OAuth2 provider
		facebookHandler := provider.CustomHandlerOpt{
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://www.facebook.com/v18.0/dialog/oauth",
				TokenURL: "https://graph.facebook.com/v18.0/oauth/access_token",
			},
			InfoURL: "https://graph.facebook.com/v18.0/me?fields=id,name,email,picture",
			Scopes:  []string{"email", "public_profile"},
			MapUserFn: func(data provider.UserData, _ []byte) token.User {
				// Create a user from the Facebook data
				user := token.User{
					ID:   "facebook_" + token.HashID(sha1.New(), data.Value("id")),
					Name: data.Value("name"),
				}

				// Add email if available
				if email := data.Value("email"); email != "" {
					user.Email = email
				}

				// Add profile picture if available
				if picture := data.Value("picture"); picture != "" {
					// Facebook returns picture data in a nested structure
					var pictureData map[string]interface{}
					if err := json.Unmarshal([]byte(picture), &pictureData); err == nil {
						if data, ok := pictureData["data"].(map[string]interface{}); ok {
							if url, ok := data["url"].(string); ok {
								user.Picture = url
							}
						}
					}
				} else {
					// Fallback to the Facebook profile picture URL format
					user.Picture = fmt.Sprintf("https://graph.facebook.com/%s/picture?type=large", data.Value("id"))
				}

				// Save the user to the database
				_ = saveOrUpdateSSOUser(db, clock, user)

				return user
			},
		}

		// Add the Facebook provider with the custom handler
		service.AddCustomProvider("facebook",
			auth.Client{
				Cid:     config.Providers.Facebook.ClientID,
				Csecret: config.Providers.Facebook.ClientSecret,
			},
			facebookHandler,
		)
	}

	// Add Discord provider if enabled
	if config.Providers.Discord.Enabled {
		// Create a custom Discord OAuth2 provider
		// The callback URL is automatically generated by the auth service

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
				_ = saveOrUpdateSSOUser(db, clock, user)

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

		// Discord provider added successfully
	}

	// Add Twitter provider if enabled
	if config.Providers.Twitter.Enabled {

		// Create a custom Twitter OAuth2 provider
		twitterHandler := provider.CustomHandlerOpt{
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://twitter.com/i/oauth2/authorize",
				TokenURL: "https://api.twitter.com/2/oauth2/token",
			},
			InfoURL: "https://api.twitter.com/2/users/me?user.fields=profile_image_url,name,username,id",
			Scopes:  []string{"users.read", "tweet.read"},
			MapUserFn: func(data provider.UserData, rawData []byte) token.User {
				lgr.Printf("INFO: Twitter auth callback received. Raw data: %s", string(rawData))
				lgr.Printf("INFO: Twitter user data values: %+v", data)
				// Twitter API v2 returns data in a nested structure
				var userData map[string]interface{}

				// Try to parse the response data
				if rawData := data.Value("data"); rawData != "" {
					if err := json.Unmarshal([]byte(rawData), &userData); err != nil {
						// If we can't parse the nested data, use the top-level data
						userData = map[string]interface{}{
							"id":       data.Value("id"),
							"name":     data.Value("name"),
							"username": data.Value("username"),
							"email":    data.Value("email"),
						}
					}
				} else {
					// If there's no nested data, use the top-level data
					userData = map[string]interface{}{
						"id":       data.Value("id"),
						"name":     data.Value("name"),
						"username": data.Value("username"),
						"email":    data.Value("email"),
					}
				}

				// Extract user information from the parsed data
				id := getStringValue(userData, "id")
				name := getStringValue(userData, "name")
				username := getStringValue(userData, "username")
				email := getStringValue(userData, "email")
				profileImageURL := getStringValue(userData, "profile_image_url")

				// Create a user from the Twitter data
				user := token.User{
					ID:   "twitter_" + token.HashID(sha1.New(), id),
					Name: name,
				}

				// Use username if name is not available
				if user.Name == "" {
					user.Name = username
				}

				// Add email if available
				if email != "" {
					user.Email = email
				}

				// Add profile picture if available
				if profileImageURL != "" {
					// Remove "_normal" to get the full-size image
					user.Picture = strings.Replace(profileImageURL, "_normal", "", 1)
				}

				// Save the user to the database
				_ = saveOrUpdateSSOUser(db, clock, user)

				return user
			},
		}

		// Add the Twitter provider with the custom handler
		service.AddCustomProvider("twitter",
			auth.Client{
				Cid:     config.Providers.Twitter.ClientID,
				Csecret: config.Providers.Twitter.ClientSecret,
			},
			twitterHandler,
		)
	}

}

// Helper function to safely extract string values from a map
func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}
