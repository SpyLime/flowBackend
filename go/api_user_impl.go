package openapi

import (
	bolt "go.etcd.io/bbolt"
)

func postUser(db *bolt.DB, user UpdateUserRequest) (userId string, err error) {
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

}
