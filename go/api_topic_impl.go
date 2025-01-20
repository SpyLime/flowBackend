package openapi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/SpyLime/flowBackend/utility"
	bolt "go.etcd.io/bbolt"
)

func getTopics(db *bolt.DB) (response []GetTopics200ResponseInner, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getTopicsRx(tx)
		return err
	})

	return
}

func getTopicsRx(tx *bolt.Tx) (response []GetTopics200ResponseInner, err error) {
	topicsBucket := tx.Bucket([]byte(utility.KeyTopics))
	if topicsBucket == nil {
		return response, fmt.Errorf("cannot find topics bucket")
	}

	c := topicsBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		var newTopic string
		err = json.Unmarshal(k, &newTopic)
		response = append(response, GetTopics200ResponseInner{
			Title: newTopic,
		})
	}

	return
}

func PostTopic(db *bolt.DB, clock utility.Clock, topic GetTopics200ResponseInner) (response ResponsePostTopic, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		response, err = postTopicTx(tx, clock, topic)
		return err
	})

	return
}

func postTopicTx(tx *bolt.Tx, clock utility.Clock, topic GetTopics200ResponseInner) (response ResponsePostTopic, err error) {

	topicsBucket, err := tx.CreateBucketIfNotExists([]byte(utility.KeyTopics))
	if err != nil {
		return
	}

	topicBucket, err := topicsBucket.CreateBucket([]byte(topic.Title))
	if err != nil {
		return response, err
	}

	nodesBucket, err := topicBucket.CreateBucket([]byte(utility.KeyNodes))
	if err != nil {
		return
	}

	response.Topic.Title = topic.Title

	newNode := NodeData{
		Topic:     topic.Title,
		CreatedBy: "change hard code",
	}

	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	id := clock.Now().Truncate(time.Millisecond)
	newNode.Id = id

	response.NodeData = AddTopic200ResponseNodeData(newNode)

	idB, err := id.MarshalText()
	if err != nil {
		return
	}

	err = nodesBucket.Put(idB, marshal)

	return
}
