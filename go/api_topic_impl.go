package openapi

import (
	"encoding/json"
	"fmt"

	"github.com/SpyLime/flowBackend/utility"
	bolt "go.etcd.io/bbolt"
)

func PostTopic(db *bolt.DB, topic GetTopics200ResponseInner) (topicId string, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		topicId, err = postTopicTx(tx, user)
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func postTopicTx(tx *bolt.Tx, topic GetTopics200ResponseInner) (topicId string, err error) {

	topicsBucket, err := tx.CreateBucketIfNotExists([]byte(utility.KeyTopics))
	if err != nil {
		return
	}

	foundTopic := topicsBucket.Get([]byte(topic.Title))
	if foundTopic != nil {
		return topicId, fmt.Errorf("There is already a topic with that title")
	}

	//can't go forward as I am tired and I don't understand why spec says I need nodeData to return with this

	marshal, err := json.Marshal(user)
	if err != nil {
		return
	}

	err = usersBucket.Put([]byte(userId), marshal)

	return
}
