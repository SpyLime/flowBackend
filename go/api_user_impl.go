package openapi

import (
	"encoding/json"

	"github.com/SpyLime/flowBackend/utility"
	bolt "go.etcd.io/bbolt"
)

func PostUser(db *bolt.DB, user UpdateUserRequest) (userId string, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		userId, err = postUserTx(tx, user)
		if err != nil {
			return err
		}

		return nil
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
