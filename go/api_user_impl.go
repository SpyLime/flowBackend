package openapi

import (
	"encoding/json"
	"fmt"

	"github.com/SpyLime/flowBackend/utility"
	bolt "go.etcd.io/bbolt"
)

func getUser(db *bolt.DB, userId string) (response User, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getUserRx(tx, userId)
		return err
	})

	return
}

func getUserRx(tx *bolt.Tx, userId string) (response User, err error) {
	usersBucket := tx.Bucket([]byte(utility.KeyUsers))
	if usersBucket == nil {
		return response, fmt.Errorf("cannot find users bucket")
	}

	userData := usersBucket.Get([]byte(userId))
	if userData == nil {
		return response, fmt.Errorf("cannot find user data")
	}

	var user UpdateUserRequest
	err = json.Unmarshal(userData, &user)
	if err != nil {
		return
	}

	response = User(user)

	return
}

func PostUser(db *bolt.DB, user UpdateUserRequest) (userId string, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		userId, err = postUserTx(tx, user)
		return err
	})

	return
}

func postUserTx(tx *bolt.Tx, user UpdateUserRequest) (userId string, err error) {

	usersBucket, err := tx.CreateBucketIfNotExists([]byte(utility.KeyUsers))
	if err != nil {
		return
	}

	userId = user.Username + "#" + utility.RandomString(4)
	foundUser := usersBucket.Get([]byte(userId))
	for foundUser != nil {
		userId = user.Username + "#" + utility.RandomString(4)
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
