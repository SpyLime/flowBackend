package main

import (
	"encoding/json"
	"fmt"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

func updateUser(db *bolt.DB, request openapi.UpdateUserRequest) (err error) {
	// Add logging to help debug
	fmt.Printf("Updating user with ID: %s\n", request.Id)

	err = db.Update(func(tx *bolt.Tx) error {
		// Use the Id field instead of Username for consistency
		err = updateUserTx(tx, request)
		return err
	})

	return
}

func updateUserTx(tx *bolt.Tx, request openapi.UpdateUserRequest) (err error) {
	// Use the Id field for lookup
	usersBucket, user, err := getUserAndBucketRx(tx, request.Id)
	if err != nil {
		return
	}

	updateUserHelper(&user, request)

	marshal, err := json.Marshal(user)
	if err != nil {
		return
	}

	// Use the Id field for storage
	err = usersBucket.Put([]byte(request.Id), marshal)

	return
}

func getUserAndBucketRx(tx *bolt.Tx, userId string) (usersBucket *bolt.Bucket, user openapi.UpdateUserRequest, err error) {
	usersBucket = tx.Bucket([]byte(KeyUsers))
	if usersBucket == nil {
		return nil, user, fmt.Errorf("can't find users bucket")
	}

	userData := usersBucket.Get([]byte(userId))
	if userData == nil {
		// Add logging to help debug the issue
		fmt.Printf("User not found in database: %s\n", userId)
		return nil, user, fmt.Errorf("can't find user")
	}

	err = json.Unmarshal(userData, &user)
	if err != nil {
		fmt.Printf("Error unmarshaling user data: %v\n", err)
	}

	// Set the ID field explicitly
	user.Id = userId

	// Check if timestamps are zero values and set them to current time if they are
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}
	if user.LastLogin.IsZero() {
		user.LastLogin = now
	}

	return
}

func updateUserHelper(user *openapi.UpdateUserRequest, request openapi.UpdateUserRequest) {
	if request.FirstName != "" {
		user.FirstName = request.FirstName
	}
	if request.LastName != "" {
		user.LastName = request.LastName
	}
	if request.Email != "" {
		user.Email = request.Email
	}
	if request.Description != "" {
		user.Description = request.Description
	}
	if request.Location != "" {
		user.Location = request.Location
	}

	// Always update the timestamps
	user.UpdatedAt = time.Now()
	user.LastLogin = time.Now()

	user.IsFlagged = request.IsFlagged
}

func getUser(db *bolt.DB, userId string) (response openapi.User, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getUserRx(tx, userId)
		return err
	})

	return
}

func getUserRx(tx *bolt.Tx, userId string) (response openapi.User, err error) {
	usersBucket := tx.Bucket([]byte(KeyUsers))
	if usersBucket == nil {
		return response, fmt.Errorf("can't find users bucket")
	}

	userData := usersBucket.Get([]byte(userId))
	if userData == nil {
		return response, fmt.Errorf("can't find user")
	}

	var user openapi.UpdateUserRequest
	err = json.Unmarshal(userData, &user)
	if err != nil {
		return
	}

	// Ensure all nodeIds in edited nodes are properly set
	for i := range user.Edited {
		if user.Edited[i].NodeId.IsZero() {
			// If nodeId is zero time, try to find the actual node
			// This is a fallback for older data
			_, nodeData, nodeErr := nodeDataFinderTx(tx, user.Edited[i].Topic, user.Edited[i].Title)
			if nodeErr == nil {
				var node openapi.NodeData
				if jsonErr := json.Unmarshal(nodeData, &node); jsonErr == nil {
					user.Edited[i].NodeId = node.Id
				}
			}
		}
	}

	// Convert UpdateUserRequest to User
	response = openapi.User{
		Id:               userId,
		Username:         user.Username,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Email:            user.Email,
		Description:      user.Description,
		Location:         user.Location,
		Created:          user.Created,
		Edited:           user.Edited,
		BattleTestedUp:   user.BattleTestedUp,
		BattleTestedDown: user.BattleTestedDown,
		FreshUp:          user.FreshUp,
		FreshDown:        user.FreshDown,
		VideoUp:          user.VideoUp,
		VideoDown:        user.VideoDown,
		Linked:           user.Linked,
		Role:             user.Role,
		Reputation:       user.Reputation,
		IsFlagged:        user.IsFlagged,
	}

	return
}

func postUser(db *bolt.DB, user openapi.UpdateUserRequest) (userId string, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		userId, err = postUserTx(tx, user)
		return err
	})

	return
}

func postUserTx(tx *bolt.Tx, user openapi.UpdateUserRequest) (userId string, err error) {

	usersBucket, err := tx.CreateBucketIfNotExists([]byte(KeyUsers))
	if err != nil {
		return
	}

	userId = user.Username + "#" + RandomString(4)
	foundUser := usersBucket.Get([]byte(userId))
	for foundUser != nil {
		userId = user.Username + "#" + RandomString(4)
		foundUser = usersBucket.Get([]byte(userId))
	}

	user.Username = userId

	marshal, err := json.Marshal(user)
	if err != nil {
		return
	}

	err = usersBucket.Put([]byte(userId), marshal)

	return
}

func deleteUser(db *bolt.DB, userId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = deleteUserTx(tx, userId)
		return err
	})

	return
}

func deleteUserTx(tx *bolt.Tx, userId string) (err error) {

	usersBucket := tx.Bucket([]byte(KeyUsers))
	if usersBucket == nil {
		return fmt.Errorf("can't find users bucket")
	}

	err = usersBucket.Delete([]byte(userId))

	return

}
