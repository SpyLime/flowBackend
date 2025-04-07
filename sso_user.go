package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/auth/token"
	"go.etcd.io/bbolt"
)

// saveOrUpdateSSOUser saves or updates a user in the database after successful SSO authentication
func saveOrUpdateSSOUser(db *bbolt.DB, user token.User) error {
	fmt.Printf("Saving/updating SSO user in database: %+v\n", user)

	// Extract provider from the user ID (format: "provider_hash")
	parts := strings.SplitN(user.ID, "_", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid user ID format: %s", user.ID)
	}
	provider := parts[0]

	// Update the database in a transaction
	err := db.Update(func(tx *bbolt.Tx) error {
		// Get the users bucket
		b, err := tx.CreateBucketIfNotExists([]byte(KeyUsers))
		if err != nil {
			return fmt.Errorf("could not create users bucket: %w", err)
		}

		// Check if the user already exists
		existingUserBytes := b.Get([]byte(user.ID))

		if existingUserBytes != nil {
			// User exists, update
			var existingUser openapi.User
			if err := json.Unmarshal(existingUserBytes, &existingUser); err != nil {
				return fmt.Errorf("could not unmarshal existing user: %w", err)
			}

			// Update fields that might have changed
			existingUser.Username = user.Name
			existingUser.Email = user.Email
			existingUser.LastLogin = time.Now()
			existingUser.UpdatedAt = time.Now()

			// Marshal the updated user
			userBytes, err := json.Marshal(existingUser)
			if err != nil {
				return fmt.Errorf("could not marshal updated user: %w", err)
			}

			fmt.Printf("Updating existing user: %s\n", user.ID)
			return b.Put([]byte(user.ID), userBytes)
		}

		// User doesn't exist, create a new one
		newUser := openapi.User{
			Id:        user.ID,
			Username:  user.Name,
			Email:     user.Email,
			Provider:  provider,
			Role:      1, // Default role for SSO users
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LastLogin: time.Now(),
		}

		userBytes, err := json.Marshal(newUser)
		if err != nil {
			return fmt.Errorf("could not marshal new user: %w", err)
		}

		fmt.Printf("Creating new user: %s\n", user.ID)
		return b.Put([]byte(user.ID), userBytes)
	})

	if err != nil {
		fmt.Printf("Error saving/updating user: %v\n", err)
		return err
	}

	fmt.Printf("User %s saved/updated successfully\n", user.ID)
	return nil
}
