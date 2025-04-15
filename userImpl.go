package main

import (
	"encoding/json"
	"fmt"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

func updateUser(db *bolt.DB, request openapi.UpdateUserRequest) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateUserTx(tx, request)
		return err
	})

	return
}

func updateUserTx(tx *bolt.Tx, request openapi.UpdateUserRequest) (err error) {
	usersBucket, user, err := getUserAndBucketRx(tx, request.Username)
	if err != nil {
		return
	}

	updateUserHelper(&user, request)

	marshal, err := json.Marshal(user)
	if err != nil {
		return
	}

	err = usersBucket.Put([]byte(request.Username), marshal)

	return
}

func getUserAndBucketRx(tx *bolt.Tx, userId string) (usersBucket *bolt.Bucket, user openapi.UpdateUserRequest, err error) {
	usersBucket = tx.Bucket([]byte(KeyUsers))
	if usersBucket == nil {
		return nil, user, fmt.Errorf("can't find users bucket")
	}

	userData := usersBucket.Get([]byte(userId))
	if userData == nil {
		return nil, user, fmt.Errorf("can't find user")
	}

	err = json.Unmarshal(userData, &user)

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
		return response, fmt.Errorf("cannot find users bucket")
	}

	userData := usersBucket.Get([]byte(userId))
	if userData == nil {
		return response, fmt.Errorf("cannot find user data")
	}

	var user openapi.UpdateUserRequest
	err = json.Unmarshal(userData, &user)
	if err != nil {
		return response, fmt.Errorf("unmarshal error: %w", err)
	}

	user.Id = userId

	response = openapi.User(user)

	// Explicitly copy fields to ensure proper type conversion
	// response = openapi.User{
	// 	Provider:         user.Provider,
	// 	Id:               user.Id,
	// 	LastLogin:        user.LastLogin,
	// 	UpdatedAt:        user.UpdatedAt,
	// 	CreatedAt:        user.CreatedAt,
	// 	Username:         user.Username,
	// 	FirstName:        user.FirstName,
	// 	LastName:         user.LastName,
	// 	Email:            user.Email,
	// 	Role:             user.Role,
	// 	Reputation:       user.Reputation,
	// 	Description:      user.Description,
	// 	Location:         user.Location,
	// 	IsFlagged:        user.IsFlagged,
	// 	BattleTestedUp:   user.BattleTestedUp,
	// 	BattleTestedDown: user.BattleTestedDown,
	// 	FreshUp:          user.FreshUp,
	// 	FreshDown:        user.FreshDown,
	// 	Edited:           user.Edited,
	// 	Created:          user.Created,
	// 	Linked:           user.Linked, // Both types use the same type, so direct assignment should work
	// 	VideoUp:          user.VideoUp,
	// 	VideoDown:        user.VideoDown,
	// }

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
